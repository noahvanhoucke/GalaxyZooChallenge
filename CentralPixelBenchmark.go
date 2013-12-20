/* CentralPixelBenchmark.go

Central pixel benchmark for Kaggle's Galaxy Zoo competition.
This benchmark reads the center pixels in a 10x10 patch, averages
The RGB values, and clusters the training set by hashed
average RGB values. Test set images are then matched to clusters and 
assigned the corresponding average Class probabilities as predictions.

@Author: Joyce Noah-Vanhoucke
@Created: 20 December 2013
*/

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetImageRGB(filename string) (int, int, int, int) {
	/* Given an image file:
	   - open file
	   - finds central pixel
	   - returns RGB values of central pixel
	*/
	// open image file
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		println("Error opening file: ", err)
		os.Exit(-1)
	}
	defer file.Close()

	// reader takes the data from the file and puts into a buffer; used for optimization
	reader := bufio.NewReader(file)
	image, err := jpeg.Decode(reader)
	if err != nil {
		println("Error opening image file: ", err)
		os.Exit(-1)
	}

	// find central pixel, query 10x10 patch around it, average RGB results over patch
	center_x := (image.Bounds().Max.X - image.Bounds().Min.X)/2
	center_y := (image.Bounds().Max.Y - image.Bounds().Min.Y)/2
	avR := 0
	avG := 0
	avB := 0
	avA := 0
	count := 0
	for i := center_x-5; i < center_x+5; i++ {
		for j := center_y-5; j < center_y+5; j++ {
			r,g,b,a := image.At(i,j).RGBA()
			avR += int(r)
			avG += int(g)
			avB += int(b)
			avA += int(a)
			count += 1
		}
	}
	return avR/count, avG/count, avB/count, avA/count
}

func AssignClassValues(splitrow []string) []float64 {
	/* Converts the Class probability values read from file from 
	   string to float. */

	floatrow := make([]float64, 37)
	for i := 0; i < len(splitrow)-1; i++ {
		val, err := strconv.ParseFloat(splitrow[i], 64)
		floatrow[i] = val
		if err != nil {
			println("Error converting string to float ", err)
			os.Exit(-1)
		}
	}
	return floatrow
}

func GetTrainingSolutions(filename string) (map[string][]float64, []string) {
	/* Reads the training solutions file. Returns dictionary with GalaxyId
	   as the key and values as an array of float presenting the Class probability values. */

	trainingSolutions := make(map[string][]float64)
	trainFile, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		println("Error opening file: ", err)
		os.Exit(-1)
	}
	defer trainFile.Close()
	scanner := bufio.NewScanner(trainFile)
	scanner.Scan()
	header := scanner.Text()
	for scanner.Scan() {
		row := scanner.Text()
		splitrow := strings.Split(row, ",")
		values := AssignClassValues(splitrow[1:len(splitrow)])
		trainingSolutions[splitrow[0]] = values
	}
	headerRow := strings.Split(header, ",")
	return trainingSolutions, headerRow
}

func GetGalaxyRGB(filepathname string) map[string][3]int {
	/* Use Glob to get a list of all files in the images directory.
	   Return map with key = galaxyId, value = RGB values of central pixel. */

	listFiles, err := filepath.Glob(filepathname)
	if err != nil {
		println("Error using Glob")
		os.Exit(-1)
	}

	// Get the RGB values for the central pixel for each image
	println("GetGalaxyRGB(): Processing file ", filepathname)
	galaxyRGB := make(map[string][3]int)
	for i := 0; i < len(listFiles); i++ {
		galaxyId := strings.Split(listFiles[i], "\\")[2]
		galaxyId = strings.Split(galaxyId, ".")[0]
		r, g, b, _ := GetImageRGB(listFiles[i])
		// Print to stdout while processing files
		if i%5000 == 0 {
			fmt.Printf("i = %v, Galaxy Id = %v\n", i, galaxyId)
		}
		galaxyRGB[galaxyId] = [3]int{r, g, b}
	}

	if len(galaxyRGB) != len(listFiles) {
		println("GetGalaxyRGB(): Missing galaxy somewhere. galaxyRGB and listFiles have different lengths: ", len(galaxyRGB), len(listFiles))
		os.Exit(-1)
	}
	return galaxyRGB
}

func AverageGalaxySolutions(galaxyList []string, trainingSolutions map[string][]float64) []float64 {
	/* Average the training solutions for all galaxies in the galaxyList */

	var avSolutions []float64
	NProbabilities := len(trainingSolutions[galaxyList[0]])
	if NProbabilities != 37 {
		println("AverageGalaxySolutions(): NProbabilities is not 37, it is ", NProbabilities)
		os.Exit(-1)
	}

	// If only 1 galaxy, no average taken
	if len(galaxyList) == 1 {
		avSolutions = trainingSolutions[galaxyList[0]]
	} else {
		// Else average over all galaxies in the galaxyList
		avSolutions = trainingSolutions[galaxyList[0]]
		for i := 1; i < len(galaxyList); i++ {
			for j := 0; j < NProbabilities; j++ {
				avSolutions[j] += trainingSolutions[galaxyList[i]][j]
			}
		}
		for j := 0; j < NProbabilities; j++ {
			avSolutions[j] /= float64(len(galaxyList))
		}
	}
	return avSolutions
}

func GetGalaxyClusters(trainingGalaxyRGB map[string][3]int, hashFactor int) map[int][]string {
	/* Cluster training galaxies by central pixdel RGB values.
	Divide RGB values by the average intensity to work with just the color.
	Use a hashFactor to do a quick and dirty clustering. */

	galaxyClusters := make(map[int][]string)
	for galaxyId, RGB := range trainingGalaxyRGB {
		avIntensity := int((RGB[0] + RGB[1] + RGB[2]) / 3)
		RGB[0] /= avIntensity
		RGB[1] /= avIntensity
		RGB[2] /= avIntensity
		key := RGB[0]*hashFactor*hashFactor + RGB[1]*hashFactor + RGB[2]
		if _, ok := galaxyClusters[key]; ok {
			galaxyClusters[key] = append(galaxyClusters[key], galaxyId)
		} else {
			galaxyClusters[key] = []string{galaxyId}
		}
	}
	if len(galaxyClusters) < 1 {
		println("GetGalaxyClusters(): Created zero galaxy clusters.")
		os.Exit(-1)
	}
	return galaxyClusters
}

func GetSolutionsForGalaxyClusters(galaxyClusters map[int][]string, trainingSolutions map[string][]float64) map[int][]float64 {
	/* Averages the training solution values over each galaxy cluster.
	   Arguments:
	     - trainingSolutions: For each galaxyId, the solutions -- the 37 probability values.
	     - galaxyClusters: Central pixel has been clustered and mapped to a unqiue value.
	          map's key is the unique value; map's value is a list of galaxyId's for that cluster
	   Returns:
	     - galaxyClusterSolutions: probability values for each galaxy cluster. Is average
	     over a cluster's solutions
	*/

	galaxyClusterSolutions := make(map[int][]float64)
	for clusterId, galaxyList := range galaxyClusters {
		galaxyClusterSolutions[clusterId] = AverageGalaxySolutions(galaxyList, trainingSolutions)
	}
	return galaxyClusterSolutions
}

func BuildPredictionRow(galaxyId string, pred []float64) []string {
	/* Build the string to be written to file using the csv library. */

	var line []string
	line = append(line, galaxyId)
	for i := 0; i < len(pred); i++ {
		line = append(line, strconv.FormatFloat(pred[i], 'f', -1, 64))
	}

	if len(line) != 38 {
		println("BuildPredictionRow(): line does not have 38 elements, it has: ", len(line))
		os.Exit(-1)
	}
	return line
}

func CreateCentralPixelBenchmark(galaxyClusterSolutions map[int][]float64,
	testGalaxyRGB map[string][3]int, testPredictionFile string, hashFactor int, headerRow []string) {
	/* Central Pixel Benchmark: using the average values for solutions for galaxies clustered by the
	   central pixel. If no match, then use 0.0 for all values. */

	testPredictions := make(map[string][]float64)
	for galaxyId, RGB := range testGalaxyRGB {
		avIntensity := int((RGB[0] + RGB[1] + RGB[2]) / 3)
		rgbCluster := hashFactor*hashFactor*(RGB[0]/avIntensity) + hashFactor*(RGB[1]/avIntensity) + RGB[2]/avIntensity
		if _, ok := galaxyClusterSolutions[rgbCluster]; ok {
			testPredictions[galaxyId] = galaxyClusterSolutions[rgbCluster]
		} else {
			println("CreateCentralPixelBenchmark(): Key not found for this GalaxyId, rgbCluster = ", galaxyId, rgbCluster)
			testPredictions[galaxyId] = []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
		}
	}

	// Write test predictions to file
	fo, err := os.Create(testPredictionFile)
	if err != nil {
		println("CreateCentralPixelBenchmark(): Error opening file: ", err)
		os.Exit(-1)
	}
	defer fo.Close()
	wr := csv.NewWriter(fo)

	// write header to file
	err = wr.Write(headerRow)
	println("Lengths of testPredictions and testGalaxyRGB: ", len(testPredictions), len(testGalaxyRGB))
	count := 0
	if len(testPredictions) == len(testGalaxyRGB) {
		for galaxyId, _ := range testGalaxyRGB {
			if len(testPredictions[galaxyId]) != 37 {
				println("Do not have 37 predictions for galaxy ", galaxyId)
				os.Exit(-1)
			}
			err = wr.Write(BuildPredictionRow(galaxyId, testPredictions[galaxyId]))
			count += 1
		}
		wr.Flush()
	} else {
		println("Lengths don't match -- missing galaxy somewhere")
		os.Exit(-1)
	}
	println("Number of predictions made = ", count)
}

func main() {

	// Get RGB values for central pixel in training and test images
	trainingImagesPath := "images_training/*.jpg"
	testImagesPath := "images_test/*.jpg"
	trainingGalaxyRGB := GetGalaxyRGB(trainingImagesPath)
	testGalaxyRGB := GetGalaxyRGB(testImagesPath)

	trainingSolutions, headerRow := GetTrainingSolutions("solutions_training.csv")

	hashFactor := 10
	galaxyClusters := GetGalaxyClusters(trainingGalaxyRGB, hashFactor)
	galaxyClusterSolutions := GetSolutionsForGalaxyClusters(galaxyClusters, trainingSolutions)

	// Given test image central pixel RGB value, find matching galaxy cluster and
	// use the average galaxy cluster solution for the test image's solution
	testPredictionFile := "lastrun.csv"
	CreateCentralPixelBenchmark(galaxyClusterSolutions, testGalaxyRGB, testPredictionFile, hashFactor, headerRow)

	println("\nEnd Program. Success!")
}
