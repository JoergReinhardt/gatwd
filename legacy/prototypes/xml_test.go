package prototypes

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"testing"
)

var (
	pjd = []byte(string(`<pjd>
   <job-id>2781</job-id>
   <transaction>1</transaction>
   <server>
      <name>Povinho2756</name>
      <uuid>98e0ca54-92b0-9c52-9061-1ff75dc12631</uuid>
      <vnc-port>6937</vnc-port>
      <vnc-listen-ip>0.0.0.0</vnc-listen-ip>
      <description>Virtual Server: 2756</description>
      <ram-in-mb>2048</ram-in-mb>
      <cpu-units>4</cpu-units>
      <cpu-set>0,1,2,3</cpu-set>
      <images>
         <image>
            <filename>uuid://ad87119f-7b8d-47e3-99d0-f24f5b8f2dff</filename>
            <bootorder>1</bootorder>
            <type>ide</type>
            <bus>1</bus>
            <unit>0</unit>

            <storage_uuid>ad87119f-7b8d-47e3-99d0-f24f5b8f2dff</storage_uuid>
            <readonly>yes</readonly>  <!-- optional, but must be 'yes' if used -->
            <volumes>
               <volume>
                  <storage_node>49434D53-0200-9071-2500-111111111111</storage_node>
                  <storage_guid>600144f0-d7b4-41da-a3a2-d0a051bcdf2e</storage_guid>
                  <storage_ip>192.168.49.58</storage_ip>
                  <storage_name>storage-lisboa-eth0</storage_name>
               </volume>
               <volume>
                  <storage_node>49434D53-0200-9071-2500-222222222222</storage_node>
                  <storage_guid>600144f0-d7b4-41da-a3a2-d0a051bcdf2f</storage_guid>
                  <storage_ip>10.132.0.65</storage_ip>
                  <storage_name>storage-lisboa-ib1</storage_name>
               </volume>
            </volumes>
         </image>
      </images>
      <storage>
         <filename>uuid://ad87119f-7b8d-47e3-99d0-f24f5b8f2dfe</filename>
         <type>virtio</type>
         <domain>0</domain>
         <bus>0</bus>
         <slot>5</slot>

         <storage_uuid>ad87119f-7b8d-47e3-99d0-f24f5b8f2dff</storage_uuid>
         <readonly>no</readonly>   <!-- optional, but must be 'no' if used -->
         <volumes>
            <volume pos="1">
               <storage_node>49434D53-0200-9071-2500-111111111111</storage_node>
               <storage_guid>600144f0-d7b4-41da-a3a2-d0a051bcdf2a</storage_guid>
               <storage_ip>192.168.49.58</storage_ip>
               <storage_name>storage-lisboa-eth0</storage_name>
            </volume>
            <volume pos="2">
               <storage_node>49434D53-0200-9071-2500-222222222222</storage_node>
               <storage_guid>600144f0-d7b4-41da-a3a2-d0a051bcdf2b</storage_guid>
               <storage_ip>10.132.0.65</storage_ip>
               <storage_name>storage-lisboa-ib1</storage_name>
            </volume>
         </volumes>
      </storage>
      <network-interface>
         <type>virtio</type>
         <mac>02:00:0a:c9:da:74</mac>
         <domain>0</domain>
         <bus>0</bus>
         <slot>6</slot>
      </network-interface>
   </server>
   <command>create-server</command>
</pjd>`))
	buf = bytes.NewBuffer(pjd)
	dec = xml.NewDecoder(buf)
)

func TestPrintWalk(t *testing.T) {

	var n XMLNode

	err := dec.Decode(&n)
	if err != nil {
		panic(err)
	}

	walkXML([]XMLNode{n}, func(n XMLNode) bool {
		if len(n.Nodes) == 0 {
			fmt.Printf("%s: %s\n", n.XMLName.Local, string(n.Content))
			if len(n.Attrs) > 0 {
				fmt.Println(n.Attrs)
			}
		} else {
			fmt.Println(n.XMLName.Local)
			if len(n.Attrs) > 0 {
				fmt.Println(n.Attrs)
			}
		}
		return true
	})
}
