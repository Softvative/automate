require 'api/interservice/compliance/profiles/profiles_pb'
require 'api/interservice/compliance/profiles/profiles_services_pb'

describe File.basename(__FILE__) do
  Profiles = Chef::Automate::Domain::Compliance::Profiles unless defined?(Profiles)

  def profiles;
    Profiles::ProfilesService;
  end

  it "checks for a missing profile metadata" do
    res = GRPC profiles, :meta_exists, Profiles::Sha256.new(sha256: '123456')
    assert_equal res.exists, false
  end

  it "checks for an existing profile metadata" do
    res = GRPC profiles, :meta_exists, Profiles::Sha256.new(sha256: '5596bb07ef4f11fd2e03a0a80c4adb7c61fc0b4d0aa6c1410b3c715c94b367da')
    assert_equal res.exists, true
  end

end
