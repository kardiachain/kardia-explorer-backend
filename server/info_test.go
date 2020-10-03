// Package server
package server_test

import (
	"testing"
)

func TestDB_InsertWithMassRecord(t *testing.T) {
	recordSize := []uint{1000, 5000, 10000, 15000, 20000, 25000, 30000, 35000}
	for _, size := range recordSize {
		generateRecordSet(size)
	}
}

func generateRecordSet(size uint) {

}

func TestDB_UsingPG(t *testing.T) {

}

func TestDB_UsingMgo(t *testing.T) {

}
