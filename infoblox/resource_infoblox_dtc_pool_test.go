package infoblox

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ibclient "github.com/infobloxopen/infoblox-go-client/v2"
	"github.com/infobloxopen/infoblox-go-client/v2/utils"
	"reflect"
	"regexp"
	"testing"
)

func testDtcPoolDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "resource_dtc_pool" {
			continue
		}
		connector := meta.(ibclient.IBConnector)
		objMgr := ibclient.NewObjectManager(connector, "terraform_test", "test")
		rec, _ := objMgr.GetDtcPoolByRef(rs.Primary.ID)
		if rec != nil {
			return fmt.Errorf("pool not found")
		}
	}
	return nil
}

func testAccDtcPoolCompare(
	t *testing.T,
	resPath string,
	expectedRec *ibclient.DtcPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, found := s.RootModule().Resources[resPath]
		if !found {
			return fmt.Errorf("not found: %s", resPath)
		}

		internalId := res.Primary.Attributes["internal_id"]
		if internalId == "" {
			return fmt.Errorf("ID is not set")
		}

		ref, found := res.Primary.Attributes["ref"]
		if !found {
			return fmt.Errorf("'ref' attribute is not set")
		}

		connector := testAccProvider.Meta().(ibclient.IBConnector)
		objMgr := ibclient.NewObjectManager(
			connector,
			"terraform_test",
			"test")

		dtcPool, err := objMgr.SearchObjectByAltId("DtcPool", ref, internalId, eaNameForInternalId)
		if err != nil {
			if isNotFoundError(err) {
				if expectedRec == nil {
					return nil
				}
				return fmt.Errorf("object with Terraform ID '%s' not found, but expected to exist", internalId)
			}
		}
		// Assertion of object type and error handling
		var rec *ibclient.DtcPool
		recJson, _ := json.Marshal(dtcPool)
		err = json.Unmarshal(recJson, &rec)

		if rec.Name == nil {
			return fmt.Errorf("'name' is expected to be defined but it is not ")
		}
		if *rec.Name != *expectedRec.Name {
			return fmt.Errorf("'fqdn' does not match: got '%s', expected '%s'", *rec.Name, *expectedRec.Name)
		}
		if rec.LbPreferredMethod == "" {
			return fmt.Errorf("'lb_preferred_method' is expected to be defined but it is not")
		}
		if rec.LbPreferredMethod != expectedRec.LbPreferredMethod {
			return fmt.Errorf("'lb_preferred_method' does not match: got '%s', expected '%s'", rec.LbPreferredMethod, expectedRec.LbPreferredMethod)
		}
		if rec.AutoConsolidatedMonitors != nil {
			if expectedRec.AutoConsolidatedMonitors != nil {
				if *rec.AutoConsolidatedMonitors != *expectedRec.AutoConsolidatedMonitors {
					return fmt.Errorf("'auto_consolidated_monitors' does not match: got '%v', expected '%v'", rec.AutoConsolidatedMonitors, expectedRec.AutoConsolidatedMonitors)
				}
			}
		}
		if rec.Comment != nil {
			if expectedRec.Comment == nil {
				return fmt.Errorf("'comment' is expected to be undefined but it is not")
			}
			if *rec.Comment != *expectedRec.Comment {
				return fmt.Errorf(
					"'comment' does not match: got '%s', expected '%s'",
					*rec.Comment, *expectedRec.Comment)
			}
		} else if expectedRec.Comment != nil {
			return fmt.Errorf("'comment' is expected to be defined but it is not")
		}
		if rec.LbPreferredTopology != nil && expectedRec.LbPreferredTopology != nil {
			var topology ibclient.DtcTopology
			err := connector.GetObject(&ibclient.DtcTopology{}, *rec.LbPreferredTopology, nil, &topology)
			if err != nil {
				return fmt.Errorf("error getting topology object: %s", *rec.LbPreferredTopology)
			}
			if *expectedRec.LbPreferredTopology != *topology.Name {
				return fmt.Errorf("the value of 'topology' field is '%s', but expected '%s'", *topology.Name, *expectedRec.LbPreferredTopology)
			}
		}
		if rec.LbAlternateTopology != nil && expectedRec.LbAlternateTopology != nil {
			var topology ibclient.DtcTopology
			err := connector.GetObject(&ibclient.DtcTopology{}, *rec.LbAlternateTopology, nil, &topology)
			if err != nil {
				return fmt.Errorf("error getting topology object: %s", *rec.LbAlternateTopology)
			}
			if *expectedRec.LbAlternateTopology != *topology.Name {
				return fmt.Errorf("the value of 'topology' field is '%s', but expected '%s'", *topology.Name, *expectedRec.LbAlternateTopology)
			}
		}
		if rec.Monitors != nil && expectedRec.Monitors != nil {
			if len(rec.Monitors) != len(expectedRec.Monitors) {
				return fmt.Errorf("the length of 'monitors' field is '%d' but expected '%d'", len(rec.Monitors), len(expectedRec.Monitors))
			}

			for i := range rec.Monitors {
				var monitorResult ibclient.DtcMonitorHttp
				err := connector.GetObject(&ibclient.DtcMonitorHttp{}, rec.Monitors[i].Ref, nil, &monitorResult)
				if err != nil {
					return fmt.Errorf("error getting monitor object: %s", rec.Monitors[i].Ref)
				}
				if monitorResult.Name != nil {
					rec.Monitors[i].Ref = *monitorResult.Name
				}
				if !reflect.DeepEqual(rec.Monitors[i].Ref, expectedRec.Monitors[i].Ref) {
					return fmt.Errorf("difference found at index %d: got '%v' but expected '%v'", i, rec.Monitors[i].Ref, expectedRec.Monitors[i].Ref)
				}
			}
		}
		if rec.Servers != nil && expectedRec.Servers != nil {
			if len(rec.Servers) != len(expectedRec.Servers) {
				return fmt.Errorf("the length of 'servers' field is '%d' but expected '%d'", len(rec.Servers), len(expectedRec.Servers))
			}

			for i := range rec.Servers {
				var serverResult ibclient.DtcServer
				err := connector.GetObject(&ibclient.DtcServer{}, rec.Servers[i].Server, nil, &serverResult)
				if err != nil {
					return fmt.Errorf("error getting server object: %s", rec.Servers[i].Server)
				}
				if serverResult.Name != nil {
					rec.Servers[i].Server = *serverResult.Name
				}
				if !reflect.DeepEqual(rec.Servers[i], expectedRec.Servers[i]) {
					return fmt.Errorf("difference found at index %d: got '%v' but expected '%v'", i, rec.Servers[i], expectedRec.Servers[i])
				}
			}
		}
		if rec.LbDynamicRatioPreferred != nil && expectedRec.LbDynamicRatioPreferred != nil {
			var monitorResult ibclient.DtcMonitorHttp
			err := connector.GetObject(&ibclient.DtcMonitorHttp{}, rec.LbDynamicRatioPreferred.Monitor, nil, &monitorResult)
			if err != nil {
				return fmt.Errorf("error getting monitor object: %s", rec.LbDynamicRatioPreferred.Monitor)
			}
			if monitorResult.Name != nil {
				rec.LbDynamicRatioPreferred.Monitor = *monitorResult.Name
			}
			if !reflect.DeepEqual(rec.LbDynamicRatioPreferred, expectedRec.LbDynamicRatioPreferred) {
				return fmt.Errorf(
					"the value of 'lb_dynamic_preferred' field is '%v', but expected '%v'",
					rec.LbDynamicRatioPreferred, expectedRec.LbDynamicRatioPreferred)
			}
		}
		if rec.LbDynamicRatioAlternate != nil && expectedRec.LbDynamicRatioAlternate != nil {
			var monitorResult ibclient.DtcMonitorHttp
			err := connector.GetObject(&ibclient.DtcMonitorHttp{}, rec.LbDynamicRatioAlternate.Monitor, nil, &monitorResult)
			if err != nil {
				return fmt.Errorf("error getting monitor object: %s", rec.LbDynamicRatioAlternate.Monitor)
			}
			if monitorResult.Name != nil {
				rec.LbDynamicRatioAlternate.Monitor = *monitorResult.Name
			}
			if !reflect.DeepEqual(rec.LbDynamicRatioAlternate, expectedRec.LbDynamicRatioAlternate) {
				return fmt.Errorf(
					"the value of 'lb_dynamic_alternate' field is '%v', but expected '%v'",
					rec.LbDynamicRatioAlternate, expectedRec.LbDynamicRatioAlternate)
			}
		}
		if rec.UseTtl != nil {
			if expectedRec.UseTtl == nil {
				return fmt.Errorf("'use_ttl' is expected to be undefined but it is not")
			}
			if *rec.UseTtl != *expectedRec.UseTtl {
				return fmt.Errorf(
					"'use_ttl' does not match: got '%t', expected '%t'",
					*rec.UseTtl, *expectedRec.UseTtl)
			}
			if *rec.UseTtl {
				if *rec.Ttl != *expectedRec.Ttl {
					return fmt.Errorf(
						"'TTL' usage does not match: got '%d', expected '%d'",
						rec.Ttl, expectedRec.Ttl)
				}
			}
		}
		if rec.Quorum != nil {
			if *rec.Quorum != *expectedRec.Quorum {
				return fmt.Errorf(
					"quorum value does not match: got '%d', expected '%d'", *rec.Quorum, *expectedRec.Quorum)
			}
		}
		if rec.Availability != expectedRec.Availability {
			return fmt.Errorf(
				"availability value does not match: got '%v', expected '%v'", rec.Availability, expectedRec.Availability)
		}
		if rec.AutoConsolidatedMonitors != nil {
			for i := range rec.ConsolidatedMonitors {
				var monitorResult ibclient.DtcMonitorHttp
				err := connector.GetObject(&ibclient.DtcMonitorHttp{}, rec.ConsolidatedMonitors[i].Monitor, nil, &monitorResult)
				if err != nil {
					return fmt.Errorf("error getting monitor object: %s", rec.ConsolidatedMonitors[i].Monitor)
				}
				if monitorResult.Name != nil {
					rec.ConsolidatedMonitors[i].Monitor = *monitorResult.Name
				}
				if !reflect.DeepEqual(rec.ConsolidatedMonitors[i], expectedRec.ConsolidatedMonitors[i]) {
					return fmt.Errorf("difference found at index %d: got '%v' but expected '%v'", i, rec.ConsolidatedMonitors[i], expectedRec.ConsolidatedMonitors[i])
				}
			}
		}

		return validateEAs(rec.Ea, expectedRec.Ea)
	}
}

var regexMissingLbPreferredTopology = regexp.MustCompile("preferred topology cannot be nil when preferred load balancing method is set to TOPOLOGY")
var regexMissingMonitors = regexp.MustCompile("lb_dynamic_ratio_preferred cannot be set when no monitors are defined")

func TestAccResourceDtcPool(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDtcPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "infoblox_dtc_pool" "pool" {
							name = "dtc_pool"
							comment="pool creation"
							lb_preferred_method = "DYNAMIC_RATIO"
							}
`),
				ExpectError: regexMissingMonitors,
			},
			{
				Config: fmt.Sprintf(`resource "infoblox_dtc_pool" "pool_neg"{
									name = "dtc_pool"
							comment="pool creation"
							lb_preferred_method = "TOPOLOGY"
				}`),
				ExpectError: regexMissingLbPreferredTopology,
			},
			{
				Config: fmt.Sprintf(`
					resource "infoblox_dtc_pool" "pool1" {
						name                 = "dtc_pool34"
						comment              = "pool creation"
						lb_preferred_method  = "ROUND_ROBIN"
					}
				`),
				Check: resource.ComposeTestCheckFunc(
					testAccDtcPoolCompare(t, "infoblox_dtc_pool.pool1", &ibclient.DtcPool{
						Name:              utils.StringPtr("dtc_pool34"),
						Comment:           utils.StringPtr("pool creation"),
						LbPreferredMethod: "ROUND_ROBIN",
						Quorum:            utils.Uint32Ptr(0),
						Ttl:               utils.Uint32Ptr(0),
						UseTtl:            utils.BoolPtr(false),
						Availability:      "ALL",
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "infoblox_dtc_pool" "pool2"{
				name = "dtc_pool2"
				comment="pool creation"
				monitors{
					monitor_name = "snmp"
     				monitor_type="snmp"
				}
				lb_preferred_method= "DYNAMIC_RATIO"
				lb_dynamic_ratio_preferred=  jsonencode({
						"monitor_name"="snmp"
						"monitor_type"="snmp"
						"method"="MONITOR"
						"monitor_metric"=".1.2"
						"monitor_weighing"="PRIORITY"
						"invert_monitor_metric"=true
				})
}`),
				Check: resource.ComposeTestCheckFunc(
					testAccDtcPoolCompare(t, "infoblox_dtc_pool.pool2", &ibclient.DtcPool{
						Name:              utils.StringPtr("dtc_pool2"),
						Comment:           utils.StringPtr("pool creation"),
						LbPreferredMethod: "DYNAMIC_RATIO",
						Monitors: []*ibclient.DtcMonitorHttp{
							{
								Ref: "snmp",
							},
						},
						LbDynamicRatioPreferred: &ibclient.SettingDynamicratio{
							Monitor:             "snmp",
							Method:              "MONITOR",
							MonitorMetric:       ".1.2",
							MonitorWeighing:     "PRIORITY",
							InvertMonitorMetric: true,
						},
						Ttl:          utils.Uint32Ptr(0),
						UseTtl:       utils.BoolPtr(false),
						Quorum:       utils.Uint32Ptr(0),
						Availability: "ALL",
					})),
			},
			{
				Config: fmt.Sprintf(`
					resource "infoblox_dtc_pool" "pool3"{
					name = "dtc_pool3"
					comment = "pool creation"
					lb_preferred_method= "TOPOLOGY"
					lb_preferred_topology= "topology_ruleset1"
					servers{
    					server = "dummy-server.com"
    					ratio=3
  					}
					lb_alternate_method = "DYNAMIC_RATIO"
					lb_dynamic_ratio_alternate =jsonencode({
						"monitor_name"="snmp"
						"monitor_type"="snmp"
						"method"="MONITOR"
						"monitor_metric"=".1.2"
						"monitor_weighing"="PRIORITY"
						"invert_monitor_metric"=true
				})
				monitors{
						monitor_name = "snmp"
						monitor_type="snmp"
				}
				monitors{
						monitor_name = "http"
						monitor_type="http"
				}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccDtcPoolCompare(t, "infoblox_dtc_pool.pool3", &ibclient.DtcPool{
						Name:                utils.StringPtr("dtc_pool3"),
						Comment:             utils.StringPtr("pool creation"),
						LbPreferredMethod:   "TOPOLOGY",
						LbPreferredTopology: utils.StringPtr("topology_ruleset1"),
						LbAlternateMethod:   "DYNAMIC_RATIO",
						Monitors: []*ibclient.DtcMonitorHttp{
							{
								Ref: "snmp",
							},
							{
								Ref: "http",
							},
						},
						LbDynamicRatioAlternate: &ibclient.SettingDynamicratio{
							Monitor:             "snmp",
							Method:              "MONITOR",
							MonitorMetric:       ".1.2",
							MonitorWeighing:     "PRIORITY",
							InvertMonitorMetric: true,
						},
						Servers: []*ibclient.DtcServerLink{
							{
								Server: "dummy-server.com",
								Ratio:  3,
							},
						},
						Ttl:          utils.Uint32Ptr(0),
						UseTtl:       utils.BoolPtr(false),
						Quorum:       utils.Uint32Ptr(0),
						Availability: "ALL",
					})),
			},
			{
				Config: fmt.Sprintf(
					`resource "infoblox_dtc_pool" "pool4"{
							name = "dtc_pool4"
							comment = "pool creation"
							lb_preferred_method="TOPOLOGY"
							lb_preferred_topology= "topology_ruleset1"
							servers{
    							server = "dummy-server.com"
    							ratio=3
							}
							servers{
    							server = "dummy-server2.com"
    							ratio=4
  							}
							monitors{
								monitor_name = "snmp"
								monitor_type="snmp"
							}
							lb_alternate_method="DYNAMIC_RATIO"
							lb_dynamic_ratio_alternate =jsonencode({
								"monitor_name"="snmp"
								"monitor_type"="snmp"
								"method"="ROUND_TRIP_DELAY"
				})
						}`),
				Check: resource.ComposeTestCheckFunc(
					testAccDtcPoolCompare(t, "infoblox_dtc_pool.pool4", &ibclient.DtcPool{
						Name:                utils.StringPtr("dtc_pool4"),
						Comment:             utils.StringPtr("pool creation"),
						LbPreferredMethod:   "TOPOLOGY",
						LbPreferredTopology: utils.StringPtr("topology_ruleset1"),
						LbAlternateMethod:   "DYNAMIC_RATIO",
						LbDynamicRatioAlternate: &ibclient.SettingDynamicratio{
							Monitor:             "snmp",
							Method:              "ROUND_TRIP_DELAY",
							MonitorMetric:       "", //default values for monitor_metric , monitor_weighing and invert_monitor_metric when method is ROUND_TRIP_DELAY
							MonitorWeighing:     "RATIO",
							InvertMonitorMetric: false,
						},
						Servers: []*ibclient.DtcServerLink{
							{
								Server: "dummy-server.com",
								Ratio:  3,
							},
							{
								Server: "dummy-server2.com",
								Ratio:  4,
							},
						},
						Monitors: []*ibclient.DtcMonitorHttp{
							{
								Ref: "snmp",
							},
						},
						Ttl:          utils.Uint32Ptr(0),
						UseTtl:       utils.BoolPtr(false),
						Quorum:       utils.Uint32Ptr(0),
						Availability: "ALL",
					})),
			},
			{
				Config: fmt.Sprintf(
					`resource "infoblox_dtc_pool" "pool5"{
						name = "dtc_pool5"
						comment = "pool creation"
						lb_preferred_method="ROUND_ROBIN"
						monitors{
						monitor_name = "snmp"
						monitor_type="snmp"
						}
						monitors{
						monitor_name = "http"
						monitor_type="http"
						}
						availability = "QUORUM"
						quorum = 2
						ttl = 120
						auto_consolidated_monitors= true
					}`),
				Check: resource.ComposeTestCheckFunc(
					testAccDtcPoolCompare(t, "infoblox_dtc_pool.pool5", &ibclient.DtcPool{
						Name:              utils.StringPtr("dtc_pool5"),
						Comment:           utils.StringPtr("pool creation"),
						LbPreferredMethod: "ROUND_ROBIN",
						Monitors: []*ibclient.DtcMonitorHttp{
							{
								Ref: "snmp",
							},
							{
								Ref: "http",
							},
						},
						ConsolidatedMonitors: []*ibclient.DtcPoolConsolidatedMonitorHealth{
							{
								Monitor:                 "snmp",
								Availability:            "ALL",
								FullHealthCommunication: true,
								Members:                 []string{},
							},
							{
								Monitor:                 "http",
								Availability:            "ALL",
								FullHealthCommunication: true,
								Members:                 []string{},
							},
						},
						Quorum:       utils.Uint32Ptr(2),
						Ttl:          utils.Uint32Ptr(120),
						UseTtl:       utils.BoolPtr(true),
						Availability: "QUORUM",
					})),
			},
			{
				Config: fmt.Sprintf(
					`resource "infoblox_dtc_pool" "pool6"{
						name = "dtc_pool6"
						comment = "pool creation"
						lb_preferred_method="ROUND_ROBIN"
						monitors{
						monitor_name = "snmp"
						monitor_type="snmp"
						}
						monitors{
						monitor_name = "http"
						monitor_type="http"
						}
						availability = "QUORUM"
						quorum = 2
						ttl = 120
						consolidated_monitors{
			        		monitor_name = "http"
         					monitor_type = "http"
							members = ["infoblox.localdomain"]
							availability= "ALL"
							full_health_communication= true
					       }
						disable= true
					}`),
				Check: resource.ComposeTestCheckFunc(
					testAccDtcPoolCompare(t, "infoblox_dtc_pool.pool6", &ibclient.DtcPool{
						Name:              utils.StringPtr("dtc_pool6"),
						Comment:           utils.StringPtr("pool creation"),
						LbPreferredMethod: "ROUND_ROBIN",
						Monitors: []*ibclient.DtcMonitorHttp{
							{
								Ref: "snmp",
							},
							{
								Ref: "http",
							},
						},
						ConsolidatedMonitors: []*ibclient.DtcPoolConsolidatedMonitorHealth{
							{
								Monitor:                 "http",
								Availability:            "ALL",
								FullHealthCommunication: true,
								Members:                 []string{"infoblox.localdomain"},
							},
						},
						Quorum:       utils.Uint32Ptr(2),
						Ttl:          utils.Uint32Ptr(120),
						UseTtl:       utils.BoolPtr(true),
						Availability: "QUORUM",
						Disable:      utils.BoolPtr(true),
					})),
			},
		},
	})
}
