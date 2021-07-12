package word2vec

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestCos(c *C) {
	// aa := embeddings.getWordEmbeddingCombine2([]string{"women", "king"}, []string{"man"})
	// _, resultString := embeddings.compareEmbeddingAll(aa)
	// c.Assert(resultString[7], Equals, "queen")

	result := CosineSimilarity([]float32{1, 1, 1, 1, 1}, []float32{1, 1, 1, 1, 1})
	c.Assert(result, Equals, float32(5.0))

	result = CosineSimilarity([]float32{2, 2, 2, 2, 2}, []float32{1, 1, 1, 1, 1})
	c.Assert(result, Equals, float32(10.0))

}

func (s *MySuite) TestMerge(c *C) {
	_, result := MergeSort([]float32{0.5, 1.6, 0.2, 0.7, 2.2}, []int{0, 1, 2, 3, 4})
	for idx, i := range []int{2, 0, 3, 1, 4} {
		c.Assert(result[idx], Equals, i)
	}

}
