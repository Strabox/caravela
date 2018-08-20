package resources

import (
	"github.com/strabox/caravela/node/common/guid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMapping(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := range resMap.resourcesGUIDMap {
		for k := range resMap.resourcesGUIDMap[i] {
			for j := range resMap.resourcesGUIDMap[i][k] {
				currRange := resMap.resourcesGUIDMap[i][k][j]
				assert.Equal(t, testExpectedTestPartitions[i][k][j].LowerGUID().Int64(), currRange.LowerGUID().Int64(), "invalid")
				assert.Equal(t, testExpectedTestPartitions[i][k][j].HigherGUID().Int64(), currRange.HigherGUID().Int64(), "Invalid")
			}
		}
	}
}

func TestRandGUIDSearch_1(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDSearch(*NewResources(1, 256))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 0, "")
		assert.True(t, randGUID.Int64() < 8191, "")
	}
}

func TestRandGUIDSearch_2(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDSearch(*NewResources(1, 257))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 8191, "")
		assert.True(t, randGUID.Int64() < 12286, "")
	}
}

func TestRandGUIDSearch_3(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDSearch(*NewResources(1, 511))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 8191, "")
		assert.True(t, randGUID.Int64() < 12286, "")
	}
}

func TestRandGUIDSearch_4(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDSearch(*NewResources(1, 750))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 12286, "")
		assert.True(t, randGUID.Int64() < 16383, "")
	}
}

func TestRandGUIDSearch_5(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.RandGUIDSearch(*NewResources(1, 1512))
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.True(t, randGUID.Int64() >= 16383, "")
	assert.True(t, randGUID.Int64() < 32767, "")
}

func TestFirstGUIDOffer(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	targetResources := []Resources{
		*NewResources(1, 256),
		*NewResources(1, 257),
		*NewResources(1, 511),
		*NewResources(1, 512),
		*NewResources(1, 1024),
		*NewResources(2, 2048),
		*NewResources(2, 1512),
	}
	expected := []int64{0, 0, 0, 8191, 12286, 16383, 12286}

	for i, resources := range targetResources {
		firstGUID, err := resMap.FirstGUIDOffer(resources)
		if err != nil {
			assert.Fail(t, err.Error(), "")
		}
		assert.Equal(t, expected[i], firstGUID.Int64(), "")
	}
}

func TestLowestResources(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	lowestResources := *resMap.LowestResources()

	assert.Equal(t, *NewResources(1, 256), lowestResources, "")
}

func TestHigherRandomGUIDSearch_1(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.HigherRandGUIDSearch(*guid.NewGUIDInteger(5000), *NewResources(1, 256))
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResources(1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(12286), "")
	assert.True(t, randGUID.Int64() < int64(16383), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResources(1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(16383), "")
	assert.True(t, randGUID.Int64() < int64(32767), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResources(1, 256))
	assert.Error(t, err, "")
}

func TestHigherRandomGUIDSearch_2(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	_, err := resMap.HigherRandGUIDSearch(*guid.NewGUIDInteger(16385), *NewResources(1, 256))
	assert.Error(t, err, "")
}

func TestHigherRandomGUIDSearch_3(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.HigherRandGUIDSearch(*guid.NewGUIDInteger(12287), *NewResources(1, 550))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(16383), "")
	assert.True(t, randGUID.Int64() < int64(32767), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResources(1, 550))
	assert.Error(t, err, "")
}

func TestLowerRandomGUID_1(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(12287), *NewResources(1, 1024))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResources(1, 1024))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResources(1, 1024))
	assert.Error(t, err, "")
}

func TestLowerRandomGUIDOffer_2(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(20000), *NewResources(2, 2500))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(12286), "")
	assert.True(t, randGUID.Int64() < int64(16383), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResources(2, 2500))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResources(2, 2500))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResources(2, 2500))
	assert.Error(t, err, "")
}

func TestLowerRandomGUIDOffer_3(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(15000), *NewResources(2, 1512))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResources(2, 1512))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResources(2, 1512))
	assert.Error(t, err, "")
}

func TestLowerPartitionsOffer_1(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResources(1, 512))
	assert.NoError(t, err)

	expected := []Resources{*NewResources(1, 512), *NewResources(1, 256)}
	assert.Equal(t, res, expected, "Resources partitions does not match")
}

func TestLowerPartitionsOffer_2(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResources(1, 750))
	assert.NoError(t, err)

	expected := []Resources{*NewResources(1, 512), *NewResources(1, 256)}
	assert.Equal(t, res, expected, "Resources partitions does not match")
}

func TestLowerPartitionsOffer_3(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResources(1, 1550))
	assert.NoError(t, err)

	expected := []Resources{*NewResources(1, 1024), *NewResources(1, 512), *NewResources(1, 256)}
	assert.Equal(t, res, expected, "Resources partitions does not match")
}

func TestLowerPartitionsOffer_4(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResources(2, 1550))
	assert.NoError(t, err)

	expected := []Resources{*NewResources(1, 1024), *NewResources(1, 512), *NewResources(1, 256)}
	assert.Equal(t, res, expected, "Resources partitions does not match")
}

func TestLowerPartitionsOffer_5(t *testing.T) {
	guid.Init(16) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResources(2, 2049))
	assert.NoError(t, err)

	expected := []Resources{*NewResources(2, 2048), *NewResources(1, 1024), *NewResources(1, 512), *NewResources(1, 256)}
	assert.Equal(t, res, expected, "Resources partitions does not match")
}

var (
	testExpectedTestPartitions = [][][]guid.Range{
		{
			{
				*guid.NewGUIDRange(*guid.NewGUIDInteger(0), *guid.NewGUIDInteger(8191)),      // 0,1,256
				*guid.NewGUIDRange(*guid.NewGUIDInteger(8191), *guid.NewGUIDInteger(12286)),  // 0,1,512
				*guid.NewGUIDRange(*guid.NewGUIDInteger(12286), *guid.NewGUIDInteger(16383)), // 0,1,1024
			},
			{
				*guid.NewGUIDRange(*guid.NewGUIDInteger(16383), *guid.NewGUIDInteger(32767)), // 0,2,2048
			},
		},
		{
			{
				*guid.NewGUIDRange(*guid.NewGUIDInteger(32767), *guid.NewGUIDInteger(65535)), // 1,2,2048
			},
		},
	}

	testPartitions = &ResourcePartitions{
		cpuPowerPartitions: []CPUPowerPartition{
			{
				ResourcePartition: ResourcePartition{Value: 0, Percentage: 50},
				cpuCoresPartitions: []CPUCoresPartition{
					{
						ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
						ramPartitions: []RAMPartition{
							{ResourcePartition: ResourcePartition{Value: 256, Percentage: 50}},
							{ResourcePartition: ResourcePartition{Value: 512, Percentage: 25}},
							{ResourcePartition: ResourcePartition{Value: 1024, Percentage: 25}},
						},
					},
					{
						ResourcePartition: ResourcePartition{Value: 2, Percentage: 50},
						ramPartitions: []RAMPartition{
							{ResourcePartition: ResourcePartition{Value: 2048, Percentage: 100}},
						},
					},
				},
			},
			{
				ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
				cpuCoresPartitions: []CPUCoresPartition{
					{
						ResourcePartition: ResourcePartition{Value: 2, Percentage: 100},
						ramPartitions: []RAMPartition{
							{ResourcePartition: ResourcePartition{Value: 2048, Percentage: 100}},
						},
					},
				},
			},
		},
	}
)
