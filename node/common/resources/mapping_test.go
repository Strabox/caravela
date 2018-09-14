package resources

import (
	"github.com/strabox/caravela/node/common/guid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMapping(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
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
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDFittestSearch(*NewResourcesCPUClass(0, 1, 256))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 0, "")
		assert.True(t, randGUID.Int64() < 8191, "")
	}
}

func TestRandGUIDSearch_2(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDFittestSearch(*NewResourcesCPUClass(0, 1, 257))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 8191, "")
		assert.True(t, randGUID.Int64() < 12286, "")
	}
}

func TestRandGUIDSearch_3(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDFittestSearch(*NewResourcesCPUClass(0, 1, 511))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 8191, "")
		assert.True(t, randGUID.Int64() < 12286, "")
	}
}

func TestRandGUIDSearch_4(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	for i := 0; i < 100000; i++ {
		randGUID, err := resMap.RandGUIDFittestSearch(*NewResourcesCPUClass(0, 1, 750))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, randGUID.Int64() >= 12286, "")
		assert.True(t, randGUID.Int64() < 16383, "")
	}
}

func TestRandGUIDSearch_5(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.RandGUIDFittestSearch(*NewResourcesCPUClass(0, 1, 1512))
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.True(t, randGUID.Int64() >= 16383, "")
	assert.True(t, randGUID.Int64() < 32767, "")
}

func TestRandGUIDSearch_6(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.RandGUIDFittestSearch(*NewResourcesCPUClass(1, 2, 1512))
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.True(t, randGUID.Int64() >= 49151, "")
	assert.True(t, randGUID.Int64() < 573434, "")
}

func TestFirstGUIDOffer(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	targetResources := []Resources{
		*NewResourcesCPUClass(0, 1, 256),
		*NewResourcesCPUClass(0, 1, 257),
		*NewResourcesCPUClass(0, 1, 511),
		*NewResourcesCPUClass(0, 1, 512),
		*NewResourcesCPUClass(0, 1, 1024),
		*NewResourcesCPUClass(0, 2, 2048),
		*NewResourcesCPUClass(0, 2, 1512),
		*NewResourcesCPUClass(1, 1, 512),
		*NewResourcesCPUClass(1, 2, 1512),
		*NewResourcesCPUClass(1, 2, 3512),
		*NewResourcesCPUClass(1, 2, 8058),
	}
	expected := []int64{0, 0, 0, 8191, 12286, 16383, 12286, 8191, 32767, 49151, 57343}

	for i, resources := range targetResources {
		firstGUID, err := resMap.FirstGUIDOffer(resources)
		if err != nil {
			assert.Fail(t, err.Error(), "")
		}
		assert.Equal(t, expected[i], firstGUID.Int64(), "")
	}
}

func TestLowestResources(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	lowestResources := *resMap.LowestResources()

	assert.Equal(t, *NewResourcesCPUClass(0, 1, 256), lowestResources, "")
}

func TestHigherRandomGUIDSearch_1(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.HigherRandGUIDSearch(*guid.NewGUIDInteger(5000), *NewResourcesCPUClass(0, 1, 256))
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(12286), "")
	assert.True(t, randGUID.Int64() < int64(16383), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(16383), "")
	assert.True(t, randGUID.Int64() < int64(32767), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(32767), "")
	assert.True(t, randGUID.Int64() < int64(40959), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(40959), "")
	assert.True(t, randGUID.Int64() < int64(49151), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(49151), "")
	assert.True(t, randGUID.Int64() < int64(57343), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 256))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(57343), "")
	assert.True(t, randGUID.Int64() < int64(65535), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 256))
	assert.Error(t, err, "")
}

func TestHigherRandomGUIDSearch_2(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	_, err := resMap.HigherRandGUIDSearch(*guid.NewGUIDInteger(64000), *NewResourcesCPUClass(0, 1, 256))
	assert.Error(t, err, "")
}

func TestHigherRandomGUIDSearch_3(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.HigherRandGUIDSearch(*guid.NewGUIDInteger(12287), *NewResourcesCPUClass(0, 1, 550))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(16383), "")
	assert.True(t, randGUID.Int64() < int64(32767), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 550))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(32767), "")
	assert.True(t, randGUID.Int64() < int64(40959), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 550))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(40959), "")
	assert.True(t, randGUID.Int64() < int64(49151), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 550))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(49151), "")
	assert.True(t, randGUID.Int64() < int64(57343), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 550))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(57343), "")
	assert.True(t, randGUID.Int64() < int64(65535), "")

	randGUID, err = resMap.HigherRandGUIDSearch(*randGUID, *NewResourcesCPUClass(0, 1, 550))
	assert.Error(t, err, "")
}

func TestLowerRandomGUID_1(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(12287), *NewResourcesCPUClass(0, 1, 1024))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 1, 1024))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 1, 1024))
	assert.Error(t, err, "")
}

func TestLowerRandomGUIDOffer_2(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(20000), *NewResourcesCPUClass(0, 2, 2500))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(12286), "")
	assert.True(t, randGUID.Int64() < int64(16383), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 2500))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 2500))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 2500))
	assert.Error(t, err, "")
}

func TestLowerRandomGUIDOffer_3(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(15000), *NewResourcesCPUClass(0, 2, 1512))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 1512))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 1512))
	assert.Error(t, err, "")
}

func TestLowerRandomGUIDOffer_4(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(35000), *NewResourcesCPUClass(1, 1, 1512))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(12286), "")
	assert.True(t, randGUID.Int64() < int64(16383), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 1512))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 1512))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(0, 2, 1512))
	assert.Error(t, err, "")
}

func TestLowerRandomGUIDOffer_5(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	randGUID, err := resMap.LowerRandGUIDOffer(*guid.NewGUIDInteger(50000), *NewResourcesCPUClass(1, 2, 2049))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(40959), "")
	assert.True(t, randGUID.Int64() < int64(49151), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(1, 2, 2049))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(32767), "")
	assert.True(t, randGUID.Int64() < int64(40959), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(1, 2, 2049))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(16383), "")
	assert.True(t, randGUID.Int64() < int64(32767), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(1, 2, 2049))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(12286), "")
	assert.True(t, randGUID.Int64() < int64(16383), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(1, 2, 2049))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(8191), "")
	assert.True(t, randGUID.Int64() < int64(12286), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(1, 2, 2049))
	if err != nil {
		assert.Fail(t, err.Error(), "")
	}
	assert.True(t, randGUID.Int64() >= int64(0), "")
	assert.True(t, randGUID.Int64() < int64(8191), "")

	randGUID, err = resMap.LowerRandGUIDOffer(*randGUID, *NewResourcesCPUClass(1, 2, 2049))
	assert.Error(t, err, "")
}

func TestLowerPartitionsOffer_1(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResourcesCPUClass(0, 1, 512))
	assert.NoError(t, err)

	expected := []Resources{*NewResourcesCPUClass(0, 1, 512), *NewResourcesCPUClass(0, 1, 256)}
	assert.Equal(t, expected, res, "FreeResources partitions does not match")
}

func TestLowerPartitionsOffer_2(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResourcesCPUClass(0, 1, 750))
	assert.NoError(t, err)

	expected := []Resources{*NewResourcesCPUClass(0, 1, 512), *NewResourcesCPUClass(0, 1, 256)}
	assert.Equal(t, expected, res, "FreeResources partitions does not match")
}

func TestLowerPartitionsOffer_3(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResourcesCPUClass(0, 1, 1550))
	assert.NoError(t, err)

	expected := []Resources{*NewResourcesCPUClass(0, 1, 1024), *NewResourcesCPUClass(0, 1, 512), *NewResourcesCPUClass(0, 1, 256)}
	assert.Equal(t, expected, res, "FreeResources partitions does not match")
}

func TestLowerPartitionsOffer_4(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResourcesCPUClass(0, 2, 1550))
	assert.NoError(t, err)

	expected := []Resources{*NewResourcesCPUClass(0, 1, 1024), *NewResourcesCPUClass(0, 1, 512), *NewResourcesCPUClass(0, 1, 256)}
	assert.Equal(t, expected, res, "FreeResources partitions does not match")
}

func TestLowerPartitionsOffer_5(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResourcesCPUClass(0, 2, 2049))
	assert.NoError(t, err)

	expected := []Resources{*NewResourcesCPUClass(0, 2, 2048), *NewResourcesCPUClass(0, 1, 1024), *NewResourcesCPUClass(0, 1, 512), *NewResourcesCPUClass(0, 1, 256)}
	assert.Equal(t, expected, res, "FreeResources partitions does not match")
}

func TestLowerPartitionsOffer_6(t *testing.T) {
	guid.Init(16, 1, 1) // Use 16-bit GUID to be easily tested
	resMap := NewResourcesMap(testPartitions)

	res, err := resMap.LowerPartitionsOffer(*NewResourcesCPUClass(1, 2, 2049))
	assert.NoError(t, err)

	expected := []Resources{
		*NewResourcesCPUClass(1, 2, 2048),
		*NewResourcesCPUClass(1, 1, 2048),
		*NewResourcesCPUClass(1, 1, 1048),
		*NewResourcesCPUClass(0, 2, 2048),
		*NewResourcesCPUClass(0, 1, 1024),
		*NewResourcesCPUClass(0, 1, 512),
		*NewResourcesCPUClass(0, 1, 256)}

	assert.Equal(t, expected, res, "FreeResources partitions does not match")
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
				*guid.NewGUIDRange(*guid.NewGUIDInteger(32767), *guid.NewGUIDInteger(40959)), // 1,1,1048
				*guid.NewGUIDRange(*guid.NewGUIDInteger(40959), *guid.NewGUIDInteger(49151)), // 1,1,2048
			},
			{
				*guid.NewGUIDRange(*guid.NewGUIDInteger(49151), *guid.NewGUIDInteger(57343)), // 1,2,2048
				*guid.NewGUIDRange(*guid.NewGUIDInteger(57343), *guid.NewGUIDInteger(65535)), // 1,2,4048
			},
		},
	}

	testPartitions = &ResourcePartitions{
		cpuClassPartitions: []CPUClassPartition{
			{
				ResourcePartition: ResourcePartition{Value: 0, Percentage: 50},
				cpuCoresPartitions: []CPUCoresPartition{
					{
						ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
						memoryPartitions: []MemoryPartition{
							{ResourcePartition: ResourcePartition{Value: 256, Percentage: 50}},
							{ResourcePartition: ResourcePartition{Value: 512, Percentage: 25}},
							{ResourcePartition: ResourcePartition{Value: 1024, Percentage: 25}},
						},
					},
					{
						ResourcePartition: ResourcePartition{Value: 2, Percentage: 50},
						memoryPartitions: []MemoryPartition{
							{ResourcePartition: ResourcePartition{Value: 2048, Percentage: 100}},
						},
					},
				},
			},
			{
				ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
				cpuCoresPartitions: []CPUCoresPartition{
					{
						ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
						memoryPartitions: []MemoryPartition{
							{ResourcePartition: ResourcePartition{Value: 1048, Percentage: 50}},
							{ResourcePartition: ResourcePartition{Value: 2048, Percentage: 50}},
						},
					},
					{
						ResourcePartition: ResourcePartition{Value: 2, Percentage: 50},
						memoryPartitions: []MemoryPartition{
							{ResourcePartition: ResourcePartition{Value: 2048, Percentage: 50}},
							{ResourcePartition: ResourcePartition{Value: 4048, Percentage: 50}},
						},
					},
				},
			},
		},
	}
)
