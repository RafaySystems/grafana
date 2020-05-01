package cloudwatch

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/grafana/grafana/pkg/components/securejsondata"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/tsdb"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestCloudWatchMetrics(t *testing.T) {
	ds := mockDatasource()

	Convey("When calling getMetricsForCustomMetrics", t, func() {
		executor := &CloudWatchExecutor{
			DataSource:                 ds,
			customMetricsMetricsMap:    make(map[string]map[string]map[string]*CustomMetricsCache),
			customMetricsDimensionsMap: make(map[string]map[string]map[string]*CustomMetricsCache),
			clients: &mockClients{
				cloudWatch: mockedCloudWatch{
					Resp: cloudwatch.ListMetricsOutput{
						Metrics: []*cloudwatch.Metric{
							{
								MetricName: aws.String("Test_MetricName"),
								Dimensions: []*cloudwatch.Dimension{
									{
										Name: aws.String("Test_DimensionName"),
									},
								},
							},
						},
					},
				},
			},
		}
		metrics, _ := executor.getMetricsForCustomMetrics("us-east-1")

		Convey("Should contain Test_MetricName", func() {
			So(metrics, ShouldContain, "Test_MetricName")
		})
	})

	Convey("When calling getDimensionsForCustomMetrics", t, func() {
		executor := &CloudWatchExecutor{
			DataSource:                 ds,
			customMetricsMetricsMap:    make(map[string]map[string]map[string]*CustomMetricsCache),
			customMetricsDimensionsMap: make(map[string]map[string]map[string]*CustomMetricsCache),
			clients: &mockClients{
				cloudWatch: mockedCloudWatch{
					Resp: cloudwatch.ListMetricsOutput{
						Metrics: []*cloudwatch.Metric{
							{
								MetricName: aws.String("Test_MetricName"),
								Dimensions: []*cloudwatch.Dimension{
									{
										Name: aws.String("Test_DimensionName"),
									},
								},
							},
						},
					},
				},
			},
		}
		dimensionKeys, _ := executor.getDimensionsForCustomMetrics("us-east-1")

		Convey("Should contain Test_DimensionName", func() {
			So(dimensionKeys, ShouldContain, "Test_DimensionName")
		})
	})

	Convey("When calling handleGetRegions", t, func() {
		executor := &CloudWatchExecutor{
			DataSource:                 ds,
			customMetricsMetricsMap:    make(map[string]map[string]map[string]*CustomMetricsCache),
			customMetricsDimensionsMap: make(map[string]map[string]map[string]*CustomMetricsCache),
			clients: &mockClients{
				ec2: mockedEc2{RespRegions: ec2.DescribeRegionsOutput{
					Regions: []*ec2.Region{
						{
							RegionName: aws.String("ap-northeast-2"),
						},
					},
				},
				}},
		}
		jsonData := simplejson.New()
		jsonData.Set("defaultRegion", "default")
		executor.DataSource = &models.DataSource{
			JsonData:       jsonData,
			SecureJsonData: securejsondata.SecureJsonData{},
		}

		result, _ := executor.handleGetRegions(context.Background(), simplejson.New(), &tsdb.TsdbQuery{})

		Convey("Should return regions", func() {
			So(result[0].Text, ShouldEqual, "ap-east-1")
			So(result[1].Text, ShouldEqual, "ap-northeast-1")
			So(result[2].Text, ShouldEqual, "ap-northeast-2")
		})
	})

	Convey("When calling handleGetEc2InstanceAttribute", t, func() {
		executor := &CloudWatchExecutor{
			DataSource:                 ds,
			customMetricsMetricsMap:    make(map[string]map[string]map[string]*CustomMetricsCache),
			customMetricsDimensionsMap: make(map[string]map[string]map[string]*CustomMetricsCache),
			clients: &mockClients{
				ec2: mockedEc2{Resp: ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{
						{
							Instances: []*ec2.Instance{
								{
									InstanceId: aws.String("i-12345678"),
									Tags: []*ec2.Tag{
										{
											Key:   aws.String("Environment"),
											Value: aws.String("production"),
										},
									},
								},
							},
						},
					},
				},
				}},
		}

		json := simplejson.New()
		json.Set("region", "us-east-1")
		json.Set("attributeName", "InstanceId")
		filters := make(map[string]interface{})
		filters["tag:Environment"] = []string{"production"}
		json.Set("filters", filters)
		result, _ := executor.handleGetEc2InstanceAttribute(context.Background(), json, &tsdb.TsdbQuery{})

		Convey("Should equal production InstanceId", func() {
			So(result[0].Text, ShouldEqual, "i-12345678")
		})
	})

	Convey("When calling handleGetEbsVolumeIds", t, func() {

		executor := &CloudWatchExecutor{
			DataSource:                 ds,
			customMetricsMetricsMap:    make(map[string]map[string]map[string]*CustomMetricsCache),
			customMetricsDimensionsMap: make(map[string]map[string]map[string]*CustomMetricsCache),
			clients: &mockClients{
				ec2: mockedEc2{Resp: ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{
						{
							Instances: []*ec2.Instance{
								{
									InstanceId: aws.String("i-1"),
									BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-1-1")}},
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-1-2")}},
									},
								},
								{
									InstanceId: aws.String("i-2"),
									BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-2-1")}},
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-2-2")}},
									},
								},
							},
						},
						{
							Instances: []*ec2.Instance{
								{
									InstanceId: aws.String("i-3"),
									BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-3-1")}},
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-3-2")}},
									},
								},
								{
									InstanceId: aws.String("i-4"),
									BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-4-1")}},
										{Ebs: &ec2.EbsInstanceBlockDevice{VolumeId: aws.String("vol-4-2")}},
									},
								},
							},
						},
					},
				},
				}},
		}

		json := simplejson.New()
		json.Set("region", "us-east-1")
		json.Set("instanceId", "{i-1, i-2, i-3, i-4}")
		result, _ := executor.handleGetEbsVolumeIds(context.Background(), json, &tsdb.TsdbQuery{})

		Convey("Should return all 8 VolumeIds", func() {
			So(len(result), ShouldEqual, 8)
			So(result[0].Text, ShouldEqual, "vol-1-1")
			So(result[1].Text, ShouldEqual, "vol-1-2")
			So(result[2].Text, ShouldEqual, "vol-2-1")
			So(result[3].Text, ShouldEqual, "vol-2-2")
			So(result[4].Text, ShouldEqual, "vol-3-1")
			So(result[5].Text, ShouldEqual, "vol-3-2")
			So(result[6].Text, ShouldEqual, "vol-4-1")
			So(result[7].Text, ShouldEqual, "vol-4-2")
		})
	})

	Convey("When calling handleGetResourceArns", t, func() {
		executor := &CloudWatchExecutor{
			DataSource:                 ds,
			customMetricsMetricsMap:    make(map[string]map[string]map[string]*CustomMetricsCache),
			customMetricsDimensionsMap: make(map[string]map[string]map[string]*CustomMetricsCache),
			clients: &mockClients{
				rgta: mockedRGTA{
					Resp: resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
							{
								ResourceARN: aws.String("arn:aws:ec2:us-east-1:123456789012:instance/i-12345678901234567"),
								Tags: []*resourcegroupstaggingapi.Tag{
									{
										Key:   aws.String("Environment"),
										Value: aws.String("production"),
									},
								},
							},
							{
								ResourceARN: aws.String("arn:aws:ec2:us-east-1:123456789012:instance/i-76543210987654321"),
								Tags: []*resourcegroupstaggingapi.Tag{
									{
										Key:   aws.String("Environment"),
										Value: aws.String("production"),
									},
								},
							},
						},
					},
				},
			},
		}

		json := simplejson.New()
		json.Set("region", "us-east-1")
		json.Set("resourceType", "ec2:instance")
		tags := make(map[string]interface{})
		tags["Environment"] = []string{"production"}
		json.Set("tags", tags)
		result, _ := executor.handleGetResourceArns(context.Background(), json, &tsdb.TsdbQuery{})

		Convey("Should return all two instances", func() {
			So(result[0].Text, ShouldEqual, "arn:aws:ec2:us-east-1:123456789012:instance/i-12345678901234567")
			So(result[0].Value, ShouldEqual, "arn:aws:ec2:us-east-1:123456789012:instance/i-12345678901234567")
			So(result[1].Text, ShouldEqual, "arn:aws:ec2:us-east-1:123456789012:instance/i-76543210987654321")
			So(result[1].Value, ShouldEqual, "arn:aws:ec2:us-east-1:123456789012:instance/i-76543210987654321")

		})
	})
}

func TestParseMultiSelectValue(t *testing.T) {
	values := parseMultiSelectValue(" i-someInstance ")
	assert.Equal(t, []string{"i-someInstance"}, values)

	values = parseMultiSelectValue("{i-05}")
	assert.Equal(t, []string{"i-05"}, values)

	values = parseMultiSelectValue(" {i-01, i-03, i-04} ")
	assert.Equal(t, []string{"i-01", "i-03", "i-04"}, values)

	values = parseMultiSelectValue("i-{01}")
	assert.Equal(t, []string{"i-{01}"}, values)

}
