package main

// credits
// codes are mostly from:
// https://github.com/danieldk/go2vec/blob/8029f60947ae/go2vec.go

// other resources used:
// https://github.com/eyaler/word2vec-slim

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	// cblas "github.com/gonum/blas/cgo"
	// "gonum.org/v1/gonum/blas"
)

func makeRange(min, max int) []int {
	arr := make([]int, max-min+1)
	for i := range arr {
		arr[i] = min + i
	}
	return arr
}

func insertSort(arr []float32, arrSort []int) ([]float32, []int) {
	len1 := len(arr)
	for i := 1; i < len1; i++ {
		temp1 := arr[i]
		tempSort := arrSort[i]
		i2 := i
		for i2 > 0 && arr[i2-1] > temp1 {
			arr[i2] = arr[i2-1]
			arrSort[i2] = arrSort[i2-1]
			i2--
		}
		arr[i2] = temp1
		arrSort[i2] = tempSort
	}
	fmt.Println(arr, arrSort)
	return arr, arrSort
}

func mergeSort(arr []float32, arrSort []int) ([]float32, []int) {
	len1 := int(len(arr))
	len2 := int(len1 / 2)
	if len1 <= 5 {
		return insertSort(arr, arrSort)
	} else {
		arr1, arrSort1 := mergeSort(arr[len2:], arrSort[len2:])
		arr2, arrSort2 := mergeSort(arr[:len2], arrSort[:len2])
		tempArr := make([]float32, len1)
		tempArrSort := make([]int, len1)
		i := 0
		for len(arr1) > 0 && len(arr2) > 0 {
			if arr1[0] < arr2[0] {
				tempArr[i] = arr1[0]
				tempArrSort[i] = arrSort1[0]
				arr1 = arr1[1:]
				arrSort1 = arrSort1[1:]
			} else {
				tempArr[i] = arr2[0]
				tempArrSort[i] = arrSort2[0]
				arr2 = arr2[1:]
				arrSort2 = arrSort2[1:]
			}
			i++
		}
		if len(arr1) == 0 {
			for j := 0; j < len(arr2); j++ {
				// fmt.Println(j, len(arr2), arr2, arr1, tempArr)
				tempArr[i] = arr2[j]
				tempArrSort[i] = arrSort2[j]
				i++
			}
		} else {
			for j := 0; j < len(arr1); j++ {
				tempArr[i] = arr1[j]
				tempArrSort[i] = arrSort1[j]
				i++
			}
		}
		return tempArr, tempArrSort
	}
}

// WordSimilarity stores the similarity of a word compared to a query word.
type WordSimilarity struct {
	Word       string
	Similarity float32
}

// Embeddings is used to store a set of word embeddings, such that common
// operations can be performed on these embeddings (such as retrieving
// similar words).
type Embeddings struct {
	// blas      blas.Float32Level2 //did not work
	matrix    []float32
	embedSize int
	indices   map[string]int
	words     []string
}

func ReadWord2VecBinary(r *bufio.Reader, normalize bool) (*Embeddings, error) {
	var nWords uint64
	if _, err := fmt.Fscanf(r, "%d", &nWords); err != nil {
		return nil, err
	}

	var vSize uint64
	if _, err := fmt.Fscanf(r, "%d", &vSize); err != nil {
		return nil, err
	}

	matrix := make([]float32, nWords*vSize)
	indices := make(map[string]int)
	words := make([]string, nWords)

	for idx := 0; idx < int(nWords); idx++ {
		word, _ := r.ReadString(' ')
		word = strings.TrimSpace(word)
		indices[word] = idx
		words[idx] = word

		start := idx * int(vSize)
		if err1 := binary.Read(r, binary.LittleEndian, matrix[start:start+int(vSize)]); err1 != nil {
			return nil, err1
		}

		if normalize {
			normalizeEmbeddings(matrix[start : start+int(vSize)])
		}
	}

	return &Embeddings{
		// blas:      cblas.Implementation{},
		matrix:    matrix,
		embedSize: int(vSize),
		indices:   indices,
		words:     words,
	}, nil
}

func readEmbeddingsOrFail(t *testing.T, filename string) *Embeddings {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	embeds, err := ReadWord2VecBinary(bufio.NewReader(f), true)
	if err != nil {
		t.Fatal(err)
	}

	return embeds
}

func (e *Embeddings) Size() int {
	return len(e.indices)
}

// Normalize an embedding using its l2-norm.
func normalizeEmbeddings(embedding []float32) {
	// Normalize
	embedLen := float32(0)
	for _, val := range embedding {
		embedLen += val * val
	}

	embedLen = float32(math.Sqrt(float64(embedLen)))

	for idx, val := range embedding {
		embedding[idx] = val / embedLen
	}
}

// EmbeddingSize returns the embedding size.
func (e *Embeddings) EmbeddingSize() int {
	return e.embedSize
}


func cosineSimilarity(vec1 []float32, vec2 []float32) float32 {
	if len(vec1) != len(vec2) {
		return 0
	}
	var inner_product float32
	for i, val := range vec1 {
		inner_product += val * vec2[i]
	}
	return inner_product
}

func (e *Embeddings) getEmbedding(idx int) []float32 {
	idx2 := e.embedSize * idx

	return e.matrix[idx2 : idx2+e.embedSize]
}

func (e *Embeddings) getWordEmbedding(word string) ([]float32, error) {
	idx, ok := e.indices[word]
	if !ok {
		return []float32{-1}, errors.New("word not found")
	}
	if word == "king" {
		fmt.Println("embed size: ", idx, idx+e.embedSize)
	}
	return e.getEmbedding(idx), nil
}

func (e *Embeddings) getWordEmbeddingSim(word string, word2 string) (float32, error) {
	idx, ok := e.indices[word]
	idx2, ok2 := e.indices[word2]
	if !ok || !ok2 {
		return -1.0, errors.New("word not found")
	}
	return cosineSimilarity(e.getEmbedding(idx), e.getEmbedding(idx2)), nil
}

func (e *Embeddings) addEmbedding(vec1 []float32, vec2 []float32, divide1 float32) []float32 {
	new_vec := make([]float32, e.embedSize)
	embedLen := e.embedSize
	for i := 0; i < embedLen; i++ {
		new_vec[i] = vec1[i] + vec2[i]/divide1
	}
	return new_vec
}

func (e *Embeddings) subtractEmbedding(vec1 []float32, vec2 []float32, divide1 float32) []float32 {
	new_vec := make([]float32, e.embedSize)
	embedLen := e.embedSize
	for i := 0; i < embedLen; i++ {
		new_vec[i] = vec1[i] - vec2[i]/divide1
	}
	return new_vec
}

func (e *Embeddings) getWordEmbeddingCombine(wordsAdd []string, wordsSubtract []string) []float32 {
	combined_vec := []float32{}
	for _, word := range wordsAdd {
		idx, ok := e.indices[word]
		if ok {
			// fmt.Println("add word: " + word)
			if len(combined_vec) == 0 {
				fmt.Println(true)
				combined_vec = e.getEmbedding(idx)
			} else {
				combined_vec = e.addEmbedding(combined_vec, e.getEmbedding(idx), 1.0)
			}
		} else {
			fmt.Println("logger: word not in embedding, " + word) // for futuer logger
		}
	}
	for _, word := range wordsSubtract {
		idx2, ok2 := e.indices[word]
		if ok2 {
			fmt.Println("subtract word: " + word)
			combined_vec = e.subtractEmbedding(combined_vec, e.getEmbedding(idx2), 1.0)
		} else {
			fmt.Println("logger: word not in embedding, " + word) // for futuer logger
		}
	}

	return combined_vec
}

func (e *Embeddings) getWordEmbeddingCombine2(wordsAdd []string, wordsSubtract []string) []float32 {
	// totalLen := float32(len(wordsAdd) + len(wordsSubtract))
	combined_vec := make([]float32, e.embedSize)
	for _, word := range wordsAdd {
		idx, ok := e.indices[word]
		if ok {
			fmt.Println("add word: " + word)
			tarEmbed := e.getEmbedding(idx)
			if len(combined_vec) == 0 {
				combined_vec := make([]float32, e.embedSize)
				for i, val := range tarEmbed {
					combined_vec[i] = val / 1
				}
			} else {
				combined_vec = e.addEmbedding(combined_vec, tarEmbed, 1)
			}
			fmt.Println(combined_vec[:3], tarEmbed[:3])
		} else {
			fmt.Println("logger: word not in embedding: " + word) // for futuer logger
		}
	}
	for _, word := range wordsSubtract {
		idx2, ok2 := e.indices[word]
		if ok2 {
			tarEmbed2 := e.getEmbedding(idx2)
			combined_vec = e.subtractEmbedding(combined_vec, tarEmbed2, 1)
			fmt.Println(combined_vec[:3], tarEmbed2[:3])

		} else {
			fmt.Println("logger: word not in embedding: " + word) // for futuer logger
		}

	}

	return combined_vec
}

func (e *Embeddings) compareEmbeddingAll(tarWordVec []float32) ([]string, []float32) {
	vecLen := len(e.words)
	wordSlice := make(map[float32]string, vecLen)
	simSlice := make([]float32, vecLen)
	i := 0
	for k, v := range e.indices { //key = word, v = idx in matrix
		tempSim := cosineSimilarity(tarWordVec, e.getEmbedding(v))
		wordSlice[tempSim] = k
		simSlice[i] = tempSim
		i++
	}
	simSliceSorted, _ := mergeSort(simSlice, makeRange(0, len(simSlice)))
	newsimSliceSorted := simSliceSorted[vecLen-10 : vecLen]
	newWordsSorted := make([]string, 10)
	for i, k := range newsimSliceSorted {
		newWordsSorted[i] = wordSlice[k]
	}
	return newWordsSorted, newsimSliceSorted
}

func getWord2Vec() *Embeddings {
	// fileLoc := "C:/Users/Fong/Desktop/GoogleNews-vectors-negative300-SLIM.bin" //word2vec bin file, credits to google
	fileLoc := "C:/Users/Fong/Desktop/GoogleNews-vectors-negative300.bin" //word2vec bin file, credits to google

	f, err := os.Open(fileLoc)
	if err != nil {
		fmt.Println("error, file not found:", fileLoc)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	normalize := true

	var nWords uint64
	if _, err2 := fmt.Fscanf(r, "%d", &nWords); err != nil {
		fmt.Println("logger: ", err2)
	}

	var vSize uint64
	if _, err3 := fmt.Fscanf(r, "%d", &vSize); err != nil {
		fmt.Println("logger: ", err3)
	}

	matrix := make([]float32, nWords*vSize)
	indices := make(map[string]int)
	words := make([]string, nWords)

	for idx := 0; idx < int(nWords); idx++ {
		word, _ := r.ReadString(' ')
		word = strings.TrimSpace(word)
		indices[word] = idx
		words[idx] = word

		start := idx * int(vSize)
		if err1 := binary.Read(r, binary.LittleEndian, matrix[start:start+int(vSize)]); err1 != nil {
			fmt.Println(err1)
		}

		if normalize {
			normalizeEmbeddings(matrix[start : start+int(vSize)])
		}
	}

	embeddings := &Embeddings{
		// blas:      cblas.Implementation{},
		matrix:    matrix,
		embedSize: int(vSize),
		indices:   indices,
		words:     words,
	}
	return embeddings
}

// codes to test out the embeddings
// fmt.Println(embeddings.getWordEmbeddingSim("spain", "europe"))
// aa := embeddings.getWordEmbeddingCombine2([]string{"women", "king"}, []string{"man"})
// embeddings.compareEmbeddingAll(aa)
// aa1, _ := embeddings.getWordEmbedding("king")
// fmt.Println(cosineSimilarity(aa, aa1))

// berlin, _ := embeddings.getWordEmbedding("berlin")
// embeddings.compareEmbeddingAll(berlin)
// bb := embeddings.getWordEmbeddingCombine([]string{"paris", "morocco"}, []string{"france"})
// embeddings.compareEmbeddingAll(bb)


