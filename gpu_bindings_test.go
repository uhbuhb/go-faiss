//+build gpu

package faiss

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFlatIndexOnGpuFunctionality(t *testing.T) {
	index, err := NewIndexFlatL2(1)
	require.Nil(t, err)

	gpuIdx, err := TransferToGpu(index)
	require.Nil(t, err)

	vectorsToAdd := []float32{1,2,3,4,5}
	err = gpuIdx.Add(vectorsToAdd)
	require.Nil(t, err)

	distances, resultIds, err := gpuIdx.Search(vectorsToAdd, 5)
	require.Nil(t, err)
	require.Equal(t, int64(len(vectorsToAdd)), gpuIdx.Ntotal())

	fmt.Println(distances, resultIds, err)
	for i := range vectorsToAdd {
		require.Equal(t, int64(i), resultIds[len(vectorsToAdd)*i])
		require.Zero(t, distances[len(vectorsToAdd)*i])
	}
	//This is necessary bc RemoveIDs isn't implemented for GPUIndexs
	cpuIdx, err := TransferToCpu(gpuIdx)
	require.Nil(t, err)
	idsSelector, err := NewIDSelectorBatch([]int64{0})
	cpuIdx.RemoveIDs(idsSelector)
	gpuIdx, err = TransferToGpu(cpuIdx)
	require.Nil(t, err)
	require.Equal(t, int64(len(vectorsToAdd)-1), gpuIdx.Ntotal())

}

func TestIndexIDMapOnGPU(t *testing.T) {
	index, err := NewIndexFlatL2(1)
	require.Nil(t, err)

	indexMap, err := NewIndexIDMap(index)
	require.Nil(t, err)

	gpuIndex, err := TransferToGpu(indexMap)
	require.Nil(t, err)

	vectorsToAdd := []float32{1,2,3,4,5}
	ids := make([]int64, len(vectorsToAdd))
	for i := 0; i < len(vectorsToAdd); i++ {
		ids[i] = int64(i)
	}

	err = gpuIndex.AddWithIDs(vectorsToAdd, ids)
	require.Nil(t, err)

	distances, resultIds, err := gpuIndex.Search(vectorsToAdd, 5)
	require.Nil(t, err)
	fmt.Println(gpuIndex.D(), gpuIndex.Ntotal())
	fmt.Println(distances, resultIds, err)
	for i := range vectorsToAdd {
		require.Equal(t, ids[i], resultIds[len(vectorsToAdd)*i])
		require.Zero(t, distances[len(vectorsToAdd)*i])
	}
}

func TestTransferToGpuAndBack(t *testing.T) {
	index, err := NewIndexFlatL2(1)
	require.Nil(t, err)

	indexMap, err := NewIndexIDMap(index)
	require.Nil(t, err)

	gpuIndex, err := TransferToGpu(indexMap)
	require.Nil(t, err)

	vectorsToAdd := []float32{1,2,4,7,11}
	ids := make([]int64, len(vectorsToAdd))
	for i := 0; i < len(vectorsToAdd); i++ {
		ids[i] = int64(i)
	}

	err = gpuIndex.AddWithIDs(vectorsToAdd, ids)
	require.Nil(t, err)

	//This is necessary bc RemoveIDs isn't implemented for GPUIndexs
	cpuIdx, err := TransferToCpu(gpuIndex)
	require.Nil(t, err)
	idsSelector, err := NewIDSelectorBatch([]int64{0})
	cpuIdx.RemoveIDs(idsSelector)
	gpuIndex, err = TransferToGpu(cpuIdx)
	require.Nil(t, err)

	require.Equal(t, int64(4), gpuIndex.Ntotal())
	distances2, resultIds2, err := gpuIndex.Search([]float32{1}, 5)
	fmt.Println(distances2, resultIds2, gpuIndex.Ntotal())
	require.Nil(t, err)
	require.Equal(t, float32(1), distances2[0])


	cpuIndex, err := TransferToCpu(gpuIndex)
	require.Nil(t, err)
	require.Equal(t, int64(4), cpuIndex.Ntotal())

	idsSelector, err = NewIDSelectorBatch([]int64{0})
	cpuIndex.RemoveIDs(idsSelector)
	distances2, resultIds2, err = cpuIndex.Search([]float32{1}, 5)
	fmt.Println(distances2, resultIds2, cpuIndex.Ntotal())
	require.Nil(t, err)
	require.Equal(t, float32(1), distances2[0])

}

