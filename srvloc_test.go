// srvloc_test.go
package srvloc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_Structures(t *testing.T) {
	t.Log("Test Structures")
	//frame, _ := hex.DecodeString("0106002c0000656e0003642a00000018736572766963653a782d68706e702d646973636f7665723a00000000")
	//buf := bytes.NewBuffer(frame)
	//var svrlocQuery svrlocQuery
	//svrlocQuery.write(&buf)
	//fmt.Printf("%+v\n", svrlocQuery)

	frame, _ := hex.DecodeString("010702910000656e000300000000028128782d68702d7665723d30312928782d68702d6d61633d3663336265353031326337382928782d68702d6e756d5f706f72743d30312928782d68702d69703d3139322e3136382e3230312e3132372928782d68702d686e3d48503132433738572928782d68702d70313d4d46473a48503b4d444c3a4f66666963656a657420363730303b434d443a50434c334755492c50434c332c504a4c2c4a5045472c50434c4d2c5552462c44572d50434c2c3830322e31312c3830322e332c4445534b4a45542c44594e3b434c533a5052494e5445523b4445533a434e353833413b4349443a4850494a5649504156323b4c45444d4449533a5553422346462343432330302c5553422330372330312330323b534e3a434e333139394b4b304b303552513b533a30333830383043343834323031303231303035613031303030303034353164303031343434313830303561343631383030363434313164303031343b5a3a303130322c30353030303030393030303030313030303030383030303030383030303030312c303630302c303730303030303030303030303030303030303030302c30623030303030303030303030303030303030303030393866653030303030303030393930633030303030303030393930663030303030303030393866652c3063302c306530303030303030303030303030303030303030302c306630303030303030303030303030303030303030302c31303030303030323030303030383030303030383030303030383030303030382c3131302c31323030302c3135302c31373030303030303030303032353030303030303030303030303030303032352c3138313b2928782d68702d677569643d36633362653539353063643529")
	buf := bytes.NewBuffer(frame)

	srvlocResponse := new(srvlocResponse)
	srvlocResponse.read(buf)
	fmt.Printf("%+v\n", srvlocResponse)

}

func Test_Response(t *testing.T) {
	dev, err := ProbeHPPrinter()
	fmt.Println(err, dev)
}
