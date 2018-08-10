package gnmitest

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/gnxi/gnmi"
	"github.com/google/link022/generated/ocstruct"
	"github.com/google/link022/testkit/common"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/grpc"
)

func fakeHandleSet(ygot.ValidatedGoStruct) error {
	return nil
}

func startMockServer(t *testing.T) (string, func()) {
	// Create the GNMI server.
	model := gnmi.NewModel([]*pb.ModelData{{
		Name:         "openconfig-access-points",
		Organization: "OpenConfig working group",
		Version:      "0.1.0",
	}},
		reflect.TypeOf((*ocstruct.Device)(nil)),
		ocstruct.SchemaTree["Device"],
		ocstruct.Unmarshal,
		ocstruct.ΛEnum)

	s, err := gnmi.NewServer(model,
		nil,
		fakeHandleSet)
	if err != nil {
		t.Fatalf("Failed to create gNMI server: %v", err)
	}

	g := grpc.NewServer()
	pb.RegisterGNMIServer(g, s)

	lis, err := net.Listen("tcp", "")
	if err != nil {
		t.Fatalf("Failed in net.Listen: %v", err)
	}

	go g.Serve(lis)
	return lis.Addr().String(), func() {
		lis.Close()
	}
}

func TestRunTest(t *testing.T) {
	// Start mock gNMI server.
	serverAddr, teardown := startMockServer(t)
	defer teardown()

	// Create gNMI client.
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Dialing to %q failed: %v", serverAddr, err)
	}
	defer conn.Close()
	client := pb.NewGNMIClient(conn)

	// Define test cases.
	testCases := []struct {
		targetTest *common.TestCase
		succeeded  bool
	}{
		{
			targetTest: &common.TestCase{
				Name: "push all configurations",
				OPs: []*common.Operation{
					{
						Type: common.OPReplace,
						Path: "/access-points/access-point[hostname=link022-pi-ap]",
						Val:  "@testfile/ap_config.json",
					},
				},
			},
			succeeded: true,
		},
		{
			targetTest: &common.TestCase{
				Name: "update radio channel",
				OPs: []*common.Operation{
					{
						Type: common.OPUpdate,
						Path: "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel",
						Val:  "11",
					},
					{
						Type: common.OPUpdate,
						Path: "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel-width",
						Val:  "20",
					},
				},
			},
			succeeded: true,
		},
		{
			targetTest: &common.TestCase{
				Name: "get radio config",
				OPs: []*common.Operation{
					{
						Type: common.OPGet,
						Path: "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel",
						Val:  "11",
					},
					{
						Type: common.OPGet,
						Path: "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel-width",
						Val:  "20",
					},
				},
			},
			succeeded: true,
		},
		{
			targetTest: &common.TestCase{
				Name: "invalid config",
				OPs: []*common.Operation{
					{
						Type: common.OPGet,
						Path: "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel",
						Val:  "11",
					},
					{
						Type: common.OPUpdate,
						Path: "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel-width",
						Val:  "20",
					},
				},
			},
			succeeded: false,
		},
	}

	// Run test cases.
	for _, testCase := range testCases {
		err := RunTest(client, testCase.targetTest, 10*time.Second, 10*time.Second)
		testResult := err == nil
		if testResult != testCase.succeeded {
			t.Errorf("[%s] incorrect test result, actual = %v [%v], expected = %v", testCase.targetTest.Name, testResult, err, testCase.succeeded)
		}
	}
}
