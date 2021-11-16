/*
A1 Nutsa Chichilidze (18131956) and Ranya El-Hwigi (18227449)
*/

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

var printing_matrices bool = false

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(1)

	test_case := get_user_input()

	var a_rows_1 int = 5
	var a_cols_1 int = 4
	var b_rows_1 int = 4
	var b_cols_1 int = 5

	var a_rows_2 int = 200
	var a_cols_2 int = 300
	var b_rows_2 int = 300
	var b_cols_2 int = 200

	var a, b [][]int

	if test_case == 1 {
		printing_matrices = true
		a = random_matrix(a_rows_1, a_cols_1)
		b = random_matrix(b_rows_1, b_cols_1)
	} else if test_case == 2 {
		a = random_matrix(a_rows_2, a_cols_2)
		b = random_matrix(b_rows_2, b_cols_2)
	} else if test_case == -1 {
		return
	}

	multiply_with_channels(a, b)
	multiply_with_waitgroups(a, b)
	multiply_with_loop_splitting(a, b)

}

/*************************************************/
/**************** solution 1 *********************/
/************ goroutines & channels **************/
/*************************************************/

func multiply_with_channels(a [][]int, b [][]int) string {
	if !multipliable(a, b) {
		return "The given matrices can not be multiplied"
	} else {
		/* initializing the resulting matrix */
		rows := len(a)
		cols := len(b[0])

		c := make([][]int, rows)
		for i := 0; i < rows; i++ {
			c[i] = make([]int, cols)
		}

		ch := make(chan bool, rows*cols)

		/* performing the calculations */
		start := time.Now()

		for i := 0; i < rows; i++ {
			go func(i int) {
				for j := 0; j < cols; j++ {
					go get_result_with_channels(a, b, i, j, c, ch)
				}
			}(i)
		}

		/* taking off the boolean values off the channel to ensure that every goroutine has finished its calculations and populated the resulting matrix */
		for i := 0; i < rows*cols; i++ {
			<-ch
		}

		elapsed := time.Since(start)

		print(c, "1")
		fmt.Printf("\nTime taken to calculate for solution 1: %s ", elapsed)
		fmt.Printf("\n_________________________________________\n\n")
	}
	return ""
}

func get_result_with_channels(a [][]int, b_col [][]int, row int, col int, c [][]int, ch chan bool) {
	/* gets the product of a row and a column and
	puts the result in the correct index of the result matrix
	puts true on the input channel to notify the program that it finished the calculation*/

	for i := 0; i < len(a[0]); i++ {
		c[row][col] += a[row][i] * b_col[i][col]
	}

	ch <- true
}

/*************************************************/
/**************** solution 2 *********************/
/************ goroutines & waitgroups **************/
/*************************************************/

func multiply_with_waitgroups(a [][]int, b [][]int) string {
	if !multipliable(a, b) {
		return "The given matrices can not be multiplied"
	} else {
		/* initializing the resulting matrix */
		rows := len(a)
		cols := len(b[0])
		c := make([][]int, rows)

		for i := 0; i < rows; i++ {
			c[i] = make([]int, cols)
		}
		/* performing the calculations */
		start := time.Now()

		var wg sync.WaitGroup

		for i := 0; i < rows; i++ {
			wg.Add(1)
			go func(i int) {
				for j := 0; j < cols; j++ {
					for k := 0; k < len(b); k++ {
						c[i][j] += a[i][k] * b[k][j]
					}
					/* waiting for all goroutines to finish calculations
					   to make sure that the entire resulting matrix is populated
					   with correct values */
				}
				wg.Done()
			}(i)
		}

		wg.Wait()

		elapsed := time.Since(start)

		print(c, "2")
		fmt.Printf("\nTime taken to calculate for solution 2: %s ", elapsed)
		fmt.Printf("\n_________________________________________\n\n")
	}
	return ""
}

/*************************************************/
/**************** solution 3 *********************/
/********* goroutines & loop splitting ***********/
/*************************************************/

func multiply_with_loop_splitting(a [][]int, b [][]int) {
	if !multipliable(a, b) {
		fmt.Println("The given matrices can not be multiplied")
	} else {
		/* extract lengths of matrices */
		aRow := len(a)
		aCol := len(a[0])
		bCol := len(b[0])

		/* initializing the resulting matrix */
		rows := aRow
		cols := bCol
		c := make([][]int, rows)
		for i := 0; i < len(c); i++ {
			c[i] = make([]int, cols)
		}

		/* begin calculations */
		//calculating number of routines to split each row * col calculation
		var numOfroutine int = (aCol / 50) + 1

		//WaitGroup to handle synchronisation
		var wg sync.WaitGroup

		//starting time to calculate multiplication duration
		start := time.Now()
		for i := 0; i < len(c); i++ { //loop through rows of result matrix
			wg.Add(1)
			go func(i int) {
				for j := 0; j < len(c[0]); j++ { //loop through cols of result matrix
					chValue := make(chan int)
					go getValue(chValue, a, b, numOfroutine, i, j)
					c[i][j] = <-chValue
				}
				wg.Done()
			}(i)
		}

		//waiting for WaitGroup to be 0
		wg.Wait()

		//get time since start
		elapsed := time.Since(start)

		//print results
		print(c, "3")
		fmt.Printf("\nTime taken to calculate for solution 3: %s ", elapsed)
		fmt.Printf("\n_________________________________________\n\n")
	}

}

func getValue(ch chan int, a [][]int, b [][]int, numOfroutine int, row_index int, col_index int) {
	//calculating split loop variables
	var bottom int = 0
	var top int
	var jump = len(a[0]) % 50

	//sets up the channels, chans is an array of channels
	chans := make([]<-chan int, numOfroutine)
	for i := range chans {
		//updating top
		top = (i * 50) + jump
		//populate the array
		chans[i] = get_split_value(a, b, row_index, col_index, bottom, top)
		//sets the bottom for next time its 0 for first time
		bottom = top + 1
	}

	// recieve values
	var total = 0
	//loop through the results of the split loops
	for i := 0; i < numOfroutine; i++ {
		msg1 := <-chans[i]
		total = total + msg1
	}
	ch <- total //pass back the total of the cell
}

// Returns channel
// multiple and return sum of the multiplies values
func get_split_value(a [][]int, b [][]int, row_index int, col_index int, inFrom int, inTo int) <-chan int {
	c := make(chan int) //channel to resturn
	//goroutine that will send message to channel
	go func() {
		var sum = 0
		//indexes are 0 to length-1 so fix inTo appropriately
		if inTo == len(a[0]) {
			inTo = len(a[0]) - 1
		}
		//loop through the row and col and multiply each cell value and get its sum
		for i := inFrom; i <= inTo; i++ {
			sum += a[row_index][i] * b[i][col_index]
		}
		//send message to channel that was returned
		c <- sum
	}()
	return c // Return the channel to the caller.
}

/*************************************************/
/**************** shared functions ***************/
/*************************************************/

func random_matrix(rows int, cols int) [][]int {
	c := make([][]int, rows)
	for i := 0; i < rows; i++ {
		c[i] = make([]int, cols)
		for j := 0; j < cols; j++ {
			c[i][j] = rand.Intn(10)
		}
	}
	return c
}

func get_user_input() int {
	var test int
	fmt.Println("Please enter the test number you would like to run from the following options:")
	fmt.Println("1. test the 3 concurrent solutions using smaller matrices. Easier to determine if output is correct.")
	fmt.Println("2. test the 3 concurrent solutions using bigger matrices. Allows us to get a better understanding of time.")
	_, err := fmt.Scanf("%d", &test)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return test
}

func print(mat [][]int, solution string) {
	if !printing_matrices {
		return
	}

	row := len(mat)
	col := len(mat[0])
	fmt.Print("The resulting matrix for solution ", solution, " is:\n\n")

	for i := 0; i < row; i++ {
		for j := 0; j < col; j++ {
			if mat[i][j] < 10 {
				fmt.Printf("%d    ", mat[i][j])
			} else if mat[i][j] < 100 {
				fmt.Printf("%d   ", mat[i][j])
			} else {
				fmt.Printf("%d  ", mat[i][j])
			}
		}
		fmt.Println("")
	}
}

func multipliable(a [][]int, b [][]int) bool {
	/* checks whether two matrices are multipliable */
	a_cols := len(a[0])
	b_rows := len(b)
	return a_cols == b_rows
}
