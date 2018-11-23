package main

import (
	"fmt"
	"github.com/CS-SI/LocalDriver/local"
	"github.com/CS-SI/LocalDriver/model"
)

func main() {
	tenant := make(map[string]interface{})
	tenant["uri"] = "qemu:///system"

	clientLocal, err := (&local.Client{}).Build(tenant)
	if err != nil {
		fmt.Println("Build failed : ", err.Error())
		return
	}

	image_ID := "9c7e752d-43da-44f2-992e-3294b2326aa4"
	template_ID := "03014bb3-9096-49a7-bf5e-0f9e440ad7c6"

	hostrequest := model.HostRequest{
		ResourceName:	"test_de_ouf",
		TemplateID: 	template_ID,
		ImageID: 		image_ID,
		PublicIP: 		true,
	}

	_, err = clientLocal.CreateHost(hostrequest)
	if err != nil {
		fmt.Println("Create Host failed : ", err.Error())
		return
	}
}


////If no uri specified libvirt will try all the available hypervisors
////"qemu:///system" --> KVM
//conn, err := libvirt.NewConnect("")
//if err != nil {
//fmt.Println("Connection failed --> shutdown\n", err.Error())
//return
//}
//defer conn.Close()
//
//domains, err := conn.ListAllDomains(0)
//if err != nil {
//fmt.Println("List all domains failed --> shutdown\n", err.Error())
//return
//}
//
//for _, domain := range domains {
//ifaces, err := domain.ListAllInterfaceAddresses(0)
//if err != nil {
//fmt.Println("Get ifaces failed --> shutdown\n", err.Error())
//return
//}
//fmt.Println(ifaces)
//}


//_, err = conn.DomainCreateXML(xml_request, 0)
//if err != nil {
//fmt.Println("DomainCreate failed --> shutdown\n", err.Error())
//return
//}

//xml_request := `
//	<domain type="kvm">
//		<name> linux-vm </name>
//		<uuid> 12fdf652-a617-4635-9916-2550beb9f32c </uuid>
//		<memory> 1000000 </memory>
//		<currentMemory> 1000000 </currentMemory>
//  		<vcpu> 1 </vcpu>
//		<os>
//			<type arch="x86_64">hvm</type>
//			<boot dev="cdrom"/>
//			<boot dev="hd"/>
//		</os>
//		<features>
//			<acpi/>
//			<apic/>
//			<vmport state="off"/>
//		</features>
//		<cpu mode="custom" match="exact">
//			<model> Skylake-Client-IBRS </model>
//		</cpu>
//		<clock offset="utc">
//			<timer name="rtc" tickpolicy="catchup"/>
//			<timer name="pit" tickpolicy="delay"/>
//			<timer name="hpet" present="no"/>
//		</clock>
//		<on_reboot>destroy</on_reboot>
//		<pm>
// 			<suspend-to-mem enabled="no"/>
//			<suspend-to-disk enabled="no"/>
//		</pm>
//		<devices>
//    		<emulator>/usr/bin/kvm-spice</emulator>
//    		<disk type="file" device="disk">
//      			<driver name="qemu" type="qcow2"/>
//      			<source file="/home/armand/LibvirtStorage/linuxconfig-vm.qcow2"/>
//      			<target dev="vda" bus="virtio"/>
//    		</disk>
//    		<disk type="file" device="cdrom">
//      			<target dev="hda" bus="ide"/>
//      			<readonly/>
//    		</disk>
//    		<controller type="usb" index="0" model="ich9-ehci1"/>
//    		<controller type="usb" index="0" model="ich9-uhci1">
//      			<master startport="0"/>
//    		</controller>
//    		<controller type="usb" index="0" model="ich9-uhci2">
//      			<master startport="2"/>
//    		</controller>
//			<controller type="usb" index="0" model="ich9-uhci3">
//      			<master startport="4"/>
//    		</controller>
//    		<interface type="network">
//      			<source network="default"/>
//      			<mac address="52:54:00:c9:da:81"/>
//     			<model type="virtio"/>
//    		</interface>
//   			<input type="tablet" bus="usb"/>
//    		<graphics type="spice" port="-1" tlsPort="-1" autoport="yes">
//      			<image compression="off"/>
//    		</graphics>
//    		<console type="pty"/>
//			<channel type="spicevmc">
//      			<target type="virtio" name="com.redhat.spice.0"/>
//    		</channel>
//    		<sound model="ich6"/>
//    		<video>
//				<model type="qxl"/>
//    		</video>
//   			<redirdev bus="usb" type="spicevmc"/>
//    		<redirdev bus="usb" type="spicevmc"/>
// 		</devices>
//	</domain>
//	`

//doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
//if err != nil {
//fmt.Println("Domains listing failed --> shutdown\n", err.Error())
//return
//}

//domcfg := &libvirtxml.Domain{}
//fmt.Printf("%d running domains:\n", len(doms))
//for _, dom := range doms {
//xmldoc, err := dom.GetXMLDesc(0)
//err = xml.Unmarshal([]byte(xmldoc), domcfg)
//name, err := dom.GetName()
//if err == nil {
//fmt.Printf("  %s\n", name)
//}
//fmt.Println("OS : ", domcfg.OS.Type)
//dom.Free()
//}


//sysinfo, err := conn.GetSysinfo(0)
//if err != nil {
//	fmt.Println("GetSysinfo failed --> shutdown\n", err.Error())
//	return
//}
//fmt.Println(sysinfo)
//storagePools, err := conn.ListAllStoragePools(0)
//if err != nil {
//	fmt.Println("GetSysinfo failed --> shutdown\n", err.Error())
//	return
//}
//for _, storagePool := range storagePools {
//	fmt.Println(storagePool.GetName())
//}
//
//
//domcfg := &libvirtxml.Domain{
//	Type: "kvm",
//	Name: "demo",
//	UUID: "8f99e332-06c4-463a-9099-330fb244e1b3",
//}
//xmldoc, err := domcfg.Marshal()
//if err != nil {
//	fmt.Println("Marshall failed --> shutdown\n", err.Error())
//	return
//}
//_, err = conn.DomainCreateXML(xmldoc, 0)
//if err != nil {
//	fmt.Println("DomainCreate failed --> shutdown\n", err.Error())
//	return
//}

//fmt.Println("XMLName : ", domcfg.XMLName)
//fmt.Println("Type : ", domcfg.Type)
//fmt.Println("ID : ", domcfg.ID)
//fmt.Println("Name : ", domcfg.Name)
//fmt.Println("UUID : ", domcfg.UUID)
//fmt.Println("GenID : ", domcfg.GenID)
//fmt.Println("Title : ", domcfg.Title)
//fmt.Println("Description : ", domcfg.Description)
//fmt.Println("Metadata : ", domcfg.Metadata)
//fmt.Println("MaximumMemory : ", domcfg.MaximumMemory)
//fmt.Println("Memory : ", domcfg.Memory)
//fmt.Println("CurrentMemory : ", domcfg.CurrentMemory)
//fmt.Println("BlockIOTune : ", domcfg.BlockIOTune)
//fmt.Println("MemoryTune : ", domcfg.MemoryTune)
//fmt.Println("MemoryBacking : ", domcfg.MemoryBacking)
//fmt.Println("VCPU : ", domcfg.VCPU)
//fmt.Println("VCPUs : ", domcfg.VCPUs)
//fmt.Println("IOThreads : ", domcfg.IOThreads)
//fmt.Println("IOThreadIDs : ", domcfg.IOThreadIDs)
//fmt.Println("CPUTune : ", domcfg.CPUTune)
//fmt.Println("NUMATune : ", domcfg.NUMATune)
//fmt.Println("Resource : ", domcfg.Resource)
//fmt.Println("SysInfo : ", domcfg.SysInfo)
//fmt.Println("Bootloader : ", domcfg.Bootloader)
//fmt.Println("BootloaderArgs : ", domcfg.BootloaderArgs)
//fmt.Println("OS : ", domcfg.OS)
//fmt.Println("IDMap : ", domcfg.IDMap)
//fmt.Println("Features : ", domcfg.Features)
//fmt.Println("CPU : ", domcfg.CPU)
//fmt.Println("Clock : ", domcfg.Clock)
//fmt.Println("OnPoweroff : ", domcfg.OnPoweroff)
//fmt.Println("OnReboot : ", domcfg.OnReboot)
//fmt.Println("OnCrash : ", domcfg.OnCrash)
//fmt.Println("PM : ", domcfg.PM)
//fmt.Println("Perf : ", domcfg.Perf)
//fmt.Println("Devices : ", domcfg.Devices)
//fmt.Println("SecLabel : ", domcfg.SecLabel)
//fmt.Println("QEMUCommandline : ", domcfg.QEMUCommandline)
//fmt.Println("LXCNamespace : ", domcfg.LXCNamespace)
//fmt.Println("VMWareDataCenterPath : ", domcfg.VMWareDataCenterPath)
//fmt.Println("KeyWrap : ", domcfg.KeyWrap)
//fmt.Println("LaunchSecurity : ", domcfg.LaunchSecurity)
