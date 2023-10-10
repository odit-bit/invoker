package ipconv

import "testing"

func Test_Uint(t *testing.T) {
	ipString := "0.0.0.0"
	ipNum, err := Uint64(ipString)
	if err != nil {
		t.Fatal(err)
	}

	if ipNum != 0 {
		t.Fatalf("wrong ip Number : %v", ipNum)
	}

	ipString = "103.111.184.0"
	ipNum, err = Uint64(ipString)
	if err != nil {
		t.Fatal(err)
	}

	if ipNum != 4294967295 {
		t.Fatalf("wrong ip Number : %v", ipNum)
	}
}

func Test_Num_to_string(t *testing.T) {
	ipNum := 0
	ipStr, _ := String(uint64(ipNum))
	if ipStr != "0.0.0.0" {
		t.Fatal("wrong ip string", ipStr)
	}
}
