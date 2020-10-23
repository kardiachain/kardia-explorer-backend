// Package server
package server_test

import (
	"testing"
)

func TestDB_InsertWithMassRecord(t *testing.T) {
	numberOfBlock := []int{1000, 5000, 10000, 15000, 20000, 25000, 30000, 35000}
	numberOfTxs := []int{1000, 3000, 5000, 10000}
	for _, blockSize := range numberOfBlock {
		for _, txSize := range numberOfTxs {
			generateRecordSet(blockSize, txSize)
		}
	}
}

func generateRecordSet(blockSize int, txSize int) {

}

func TestDB_UsingPG(t *testing.T) {

}

func TestDB_UsingMgo(t *testing.T) {

}
