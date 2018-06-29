package stunning

import (
	"encoding/binary"
	"log"
	"os"
	"testing"
)

func TestAddr(t *testing.T) {
	log.SetOutput(os.Stdout)
	i := 1
	var addr []byte
	addr = append([]byte{}, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	log.Println(i)
	log.Println(len(addr))
	binary.BigEndian.PutUint64(addr, uint64(i))
	log.Println(addr)
	addr2 := int(binary.BigEndian.Uint64(addr))
	log.Println(addr2)
	if addr2 != i {
		t.Fail()
	}
}
