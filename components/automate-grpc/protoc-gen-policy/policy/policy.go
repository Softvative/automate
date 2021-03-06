package policy

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/dave/jennifer/jen"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
	options "google.golang.org/genproto/googleapis/api/annotations"

	"github.com/chef/automate/components/automate-grpc/protoc-gen-policy/iam"
)

const (
	moduleName    = "policy"
	commentFormat = `// Code generated by protoc-gen-%s. DO NOT EDIT.
// source: %s
`
	unsupportedMethod = "UNSUPPORTED"
)

// Module generates the policy mapping calls according to the proto file annotations.
type Module struct {
	*pgs.ModuleBase
	ctx pgsgo.Context
}

type policyInfo struct {
	action, resource string
}

type policyInfoConverter func(interface{}) *policyInfo

type policyInfoValidator func(pgs.Message, *policyInfo) error

type policyBundle struct {
	pkg       string               // go package to use for generating MapMethodTo calls
	extension *proto.ExtensionDesc // proto extension to look up
	convert   policyInfoConverter
	validate  policyInfoValidator
}

var policyVersion = policyBundle{
	pkg:       "github.com/chef/automate/components/automate-gateway/api/iam/v2/policy",
	extension: iam.E_Policy,
	convert: func(x interface{}) *policyInfo {
		if y, ok := x.(*iam.PolicyInfo); ok {
			return &policyInfo{action: y.Action, resource: y.Resource}
		}
		return nil
	},
	validate: func(msg pgs.Message, pi *policyInfo) error {
		expanded, err := dummyExpand(msg, pi.resource)
		if err != nil {
			return err
		}
		// using a preserved version of the protobuf validation for IAM V2 IsAuthorized
		// which was removed (see below).
		req := &ValidateV2ResourceAndActions{
			Subjects: []string{"user:local:albertine"}, // this won't fail validation
			Resource: expanded,                         // resource and
			Action:   pi.action,                        // action could
		}
		return req.validateV2()
	},
}

type httpRule struct {
	rule *options.HttpRule
}

func (h httpRule) methodAndEndpoint() (string, string) {
	switch {
	case h.rule.GetGet() != "":
		return "GET", h.rule.GetGet()

	case h.rule.GetPut() != "":
		return "PUT", h.rule.GetPut()

	case h.rule.GetPost() != "":
		return "POST", h.rule.GetPost()

	case h.rule.GetDelete() != "":
		return "DELETE", h.rule.GetDelete()

	case h.rule.GetPatch() != "":
		return "PATCH", h.rule.GetPatch()

	default:
		return unsupportedMethod, unsupportedMethod
	}
}

// Policy returns an initialized policy.Module
func Policy() *Module { return &Module{ModuleBase: &pgs.ModuleBase{}} } // nolint: govet

// Name satisfies the pgs.Module interface
func (m *Module) Name() string { return moduleName }

func (m *Module) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.ctx = pgsgo.InitContext(c.Parameters())
}

// Execute satisfies the pgs.Module interface
func (m *Module) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	for _, f := range targets {
		// skip files with no services; also assume there isn't more than ONE
		// service in each proto file
		if n := len(f.Services()); n != 1 {
			m.Debugf("%d service(s) in %v, skipping", n, f.Name())
			continue
		}
		if len(f.Services()[0].Methods()) == 0 {
			m.Debugf("service in %v has no methods, skipping", f.Name())
			continue
		}
		m.processFile(f)
	}
	return m.Artifacts()
}

func (m *Module) processFile(f pgs.File) {
	out := bytes.Buffer{}
	err := m.applyTemplate(&out, policyVersion, f)
	if err != nil {
		m.Logf("couldn't apply template: %s", err)
		m.Fail("code generation failed")
	} else {
		generatedFileName := m.ctx.OutputPath(f).SetExt(fmt.Sprintf(".%s.go", moduleName)).String()
		m.AddGeneratorFile(generatedFileName, out.String())
	}
}

type mapMethod struct {
	policyInfo               *policyInfo
	qualifiedName            string
	httpMethod, httpEndpoint string
	method                   pgs.Method
}

func newMapMethod(
	ext *proto.ExtensionDesc,
	convert policyInfoConverter,
	meth pgs.Method,
) (*mapMethod, error) {
	mm := mapMethod{
		qualifiedName: fmt.Sprintf("/%s.%s/%s", meth.Package().ProtoName(), meth.Service().Name(), meth.Name()),
		method:        meth,
	}

	http, err := extractHTTPInfoFromMethodDescriptor(meth.Descriptor())
	if err != nil {
		return nil, err
	}
	if http != nil {
		mm.httpMethod, mm.httpEndpoint = http.methodAndEndpoint()
	}
	if mm.httpMethod == unsupportedMethod {
		return nil, fmt.Errorf("%s: HTTP method '%v' is not supported", meth.Name(), http.rule.GetPattern())
	}

	// If no http info, the method is not public so lack of policy info is NOT a failure.

	mm.policyInfo, err = extractPolicyInfoFromMethodDescriptor(ext, convert, meth.Descriptor())
	if http != nil && err != nil {
		return nil, err
	}
	if http != nil && mm.policyInfo == nil {
		return nil, fmt.Errorf("%s: policy info is missing", meth.Name())
	}

	return &mm, nil
}

func dummyExpand(m pgs.Message, resourcePattern string) (string, error) {
	paramRegexp := regexp.MustCompile(`\{[a-zA-Z]\w*\}`)
	matches := paramRegexp.FindAllString(resourcePattern, -1 /* return all substrings */)
	fieldNames := map[string]bool{}

	for _, match := range matches {
		fieldNames[match[1:len(match)-1]] = true
	}

	for _, f := range m.Fields() {
		if fieldNames[f.Name().String()] {
			delete(fieldNames, f.Name().String())
		}
	}

	if len(fieldNames) > 0 {
		msg := "unknown field name(s) in pattern:"
		for f := range fieldNames {
			msg = msg + " " + f
		}
		return "", errors.New(msg)
	}

	return paramRegexp.ReplaceAllString(resourcePattern, "dummy"), nil
}

func (m *Module) applyTemplate(buf *bytes.Buffer, pol policyBundle, pgsFile pgs.File) error {
	svcs := pgsFile.Services()
	pkgName := m.ctx.PackageName(pgsFile).String()
	importPath := m.ctx.ImportPath(pgsFile).String()
	protoFileName := pgsFile.Name().String()

	f := jen.NewFilePathName(importPath, pkgName)
	f.HeaderComment(fmt.Sprintf(commentFormat, moduleName, protoFileName))

	methods := []*mapMethod{}
	for _, svc := range svcs {
		for _, method := range svc.Methods() {
			m.Debugf("processing method: %s", method.Name())
			mm, err := newMapMethod(pol.extension, pol.convert, method)
			if err != nil {
				return err
			}
			if mm.policyInfo == nil {
				continue
			}
			if err := pol.validate(method.Input(), mm.policyInfo); err != nil {
				return err
			}
			methods = append(methods, mm)
		}
	}

	if len(methods) == 0 {
		return errors.New("no method annotations found to process")
	}

	f.Func().Id("init").Params().BlockFunc(func(g *jen.Group) {
		for _, meth := range methods {
			g.Qual(pol.pkg, "MapMethodTo").Call(
				jen.Lit(meth.qualifiedName),
				jen.Lit(meth.policyInfo.resource),
				jen.Lit(meth.policyInfo.action),
				jen.Lit(meth.httpMethod),
				jen.Lit(meth.httpEndpoint),
				m.generateExpansionFunction(pol.pkg, meth.method),
			)
		}
	})

	return f.Render(buf)
}

// Note 2018/03/06 (sr): there's no deeper reason (as far as I am aware of) to
// go to the level of descriptor.*Proto messages -- maybe PGS provides the same
// information in a more convenient way. However, this is a port of existing
// code, and I've preferred to keep this as-is for now.
func extractPolicyInfoFromMethodDescriptor(
	extn *proto.ExtensionDesc,
	convert policyInfoConverter,
	meth *descriptor.MethodDescriptorProto) (*policyInfo, error) {
	ext, _, err := extractExtension(meth, extn)
	if err != nil {
		return nil, err
	}
	opts := convert(ext)
	if opts == nil {
		return nil, fmt.Errorf("unexpected extension of type %T", ext)
	}
	return opts, nil
}

func extractHTTPInfoFromMethodDescriptor(
	meth *descriptor.MethodDescriptorProto) (*httpRule, error) {
	ext, found, err := extractExtension(meth, options.E_Http)
	if !found {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return nil, fmt.Errorf("extension is %T; want an HttpRule", ext)
	}
	return &httpRule{rule: opts}, nil
}

func extractExtension(
	meth *descriptor.MethodDescriptorProto,
	ext *proto.ExtensionDesc,
) (interface{}, bool, error) {
	if meth.Options == nil {
		return nil, false, fmt.Errorf("method %s has no options", *meth.Name)
	}
	if !proto.HasExtension(meth.Options, ext) {
		return nil, false, fmt.Errorf("method %s has no extension of type %s", *meth.Name, ext.Name)
	}
	opts, err := proto.GetExtension(meth.Options, ext)
	return opts, true, err
}

func (m *Module) generateExpansionFunction(policyPkg string, meth pgs.Method) *jen.Statement {
	return jen.Func().Params(
		jen.Id("unexpandedResource").String(),
		jen.Id("input").Id("interface{}"),
	).String().BlockFunc(func(g *jen.Group) {
		var fields []string
		for _, f := range meth.Input().Fields() {
			// As of now, we only support fields of type `string`. However,
			// `repeated string` is also reported as TYPE_STRING, but with an extra
			// label for the repeated. Since we don't support anything repeated
			// either, this check can be simple:
			if f.Type().ProtoType().Proto() != descriptor.FieldDescriptorProto_TYPE_STRING ||
				f.Type().ProtoLabel().Proto() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				continue
			}
			fields = append(fields, f.Name().String())
		}
		if len(fields) > 0 {
			g.If(
				jen.List(jen.Id("m"), jen.Id("ok")).
					Op(":=").
					Id("input.").Parens(jen.Op("*").Qual(m.maybePrefixType(meth.Input()))).Op(";").Id("ok"),
			).BlockFunc(func(g *jen.Group) {
				g.Add(generateExtractionFunction(policyPkg, fields))
			})
			g.Return(jen.Lit(""))
		} else {
			g.Return(jen.Id("unexpandedResource"))
		}
	})
}

func (m *Module) maybePrefixType(msg pgs.Message) (string, string) {
	return m.ctx.ImportPath(msg).String(), m.ctx.Name(msg).String()
}

func generateExtractionFunction(policyPkg string, fields []string) *jen.Statement {
	return jen.Return(
		jen.Qual(policyPkg, "ExpandParameterizedResource").Call(jen.Id("unexpandedResource"),
			jen.Func().Params(jen.Id("want").String()).String().Block(
				jen.Switch(jen.Id("want")).BlockFunc(func(g *jen.Group) {
					for _, f := range fields {
						g.Case(jen.Lit(f)).Block(
							jen.Return(jen.Id("m").Dot(generator.CamelCase(f))),
						)
					}
					g.Default().Block(
						jen.Return(jen.Lit("")),
					)
				}),
			),
		),
	)
}

type IsAuthorizedReq struct {
	Subjects             []string `protobuf:"bytes,1,rep,name=subjects,proto3" json:"subjects,omitempty" toml:"subjects,omitempty" mapstructure:"subjects,omitempty"`
	Resource             string   `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty" toml:"resource,omitempty" mapstructure:"resource,omitempty"`
	Action               string   `protobuf:"bytes,3,opt,name=action,proto3" json:"action,omitempty" toml:"action,omitempty" mapstructure:"action,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" toml:"-" mapstructure:"-,omitempty"`
	XXX_unrecognized     []byte   `json:"-" toml:"-" mapstructure:"-,omitempty"`
	XXX_sizecache        int32    `json:"-" toml:"-" mapstructure:"-,omitempty"`
}

// We were re-using the IAM V2 IsAuthorizedReq's Validate function to
// validate our V2 AuthZ annotations. Preserving that code here so we
// can continue to validate.
type ValidateV2ResourceAndActions struct {
	Subjects             []string `protobuf:"bytes,1,rep,name=subjects,proto3" json:"subjects,omitempty" toml:"subjects,omitempty" mapstructure:"subjects,omitempty"`
	Resource             string   `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty" toml:"resource,omitempty" mapstructure:"resource,omitempty"`
	Action               string   `protobuf:"bytes,3,opt,name=action,proto3" json:"action,omitempty" toml:"action,omitempty" mapstructure:"action,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" toml:"-" mapstructure:"-,omitempty"`
	XXX_unrecognized     []byte   `json:"-" toml:"-" mapstructure:"-,omitempty"`
	XXX_sizecache        int32    `json:"-" toml:"-" mapstructure:"-,omitempty"`
}

func (m *ValidateV2ResourceAndActions) GetSubjects() []string {
	if m != nil {
		return m.Subjects
	}
	return nil
}

func (m *ValidateV2ResourceAndActions) GetResource() string {
	if m != nil {
		return m.Resource
	}
	return ""
}

func (m *ValidateV2ResourceAndActions) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

var _IsAuthorizedReq_Subjects_Pattern = regexp.MustCompile("^(?:team|user):(?:local|ldap|saml):[^:*]+$|^token:[^:*]+$|^tls:service:[^:*]+:[^:*]+$")

var _IsAuthorizedReq_Resource_Pattern = regexp.MustCompile("^[a-z][^:*]*(?::[^:*]+)*$")

var _IsAuthorizedReq_Action_Pattern = regexp.MustCompile("^[a-z][a-zA-Z]*(?::[a-z][a-zA-Z]*){2}$")

type ValidateV2ResourceAndActionsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

func (e ValidateV2ResourceAndActionsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sIsAuthorizedReq.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

func (m *ValidateV2ResourceAndActions) validateV2() error {
	if m == nil {
		return nil
	}

	if len(m.GetSubjects()) < 1 {
		return ValidateV2ResourceAndActionsValidationError{
			field:  "Subjects",
			reason: "value must contain at least 1 item(s)",
		}
	}

	for idx, item := range m.GetSubjects() {
		_, _ = idx, item

		if !_IsAuthorizedReq_Subjects_Pattern.MatchString(item) {
			return ValidateV2ResourceAndActionsValidationError{
				field:  fmt.Sprintf("Subjects[%v]", idx),
				reason: "value does not match regex pattern \"^(?:team|user):(?:local|ldap|saml):[^:*]+$|^token:[^:*]+$|^tls:service:[^:*]+:[^:*]+$\"",
			}
		}

	}

	if !_IsAuthorizedReq_Resource_Pattern.MatchString(m.GetResource()) {
		return ValidateV2ResourceAndActionsValidationError{
			field:  "Resource",
			reason: fmt.Sprintf("'%s' does not match regex pattern \"^[a-z][^:*]*(?::[^:*]+)*$\"", m.GetResource()),
		}
	}

	if !_IsAuthorizedReq_Action_Pattern.MatchString(m.GetAction()) {
		return ValidateV2ResourceAndActionsValidationError{
			field:  "Action",
			reason: fmt.Sprintf("'%s' does not match regex pattern \"^[a-z][a-zA-Z]*(?::[a-z][a-zA-Z]*){2}$\"", m.GetAction()),
		}
	}

	return nil
}
