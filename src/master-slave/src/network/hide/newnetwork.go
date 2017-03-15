package network

import (
	"flag"
	"fmt"
	//"global"
	"network/bcast"
	"network/localip"
	"network/peers"
	"strconv"
	"time"
)

const (
	masterPort = 20079
	slavePort  = 20179
)

type IP string

type Master_msg struct {
	Address IP
	//global_liste
}

type Slave_msg struct {
	Address IP
}

func bool_to_master_chan(value bool, is_master_chan chan bool) {
	is_master_chan <- value
}

func Choose_master(is_master_chan chan bool, IP_adresses peers.PeerUpdate) {
	highest_ip := 0
	str_localIP, _ := localip.LocalIP()
	fmt.Println("my ip is :", str_localIP)
	int_localIP, _ := strconv.Atoi(str_localIP[12:])
	for i := 0; i < len(IP_adresses.Peers); i++ {
		str_ip := IP_adresses.Peers[i]
		int_ip, _ := strconv.Atoi(str_ip[12:])
		if int_ip > highest_ip {
			highest_ip = int_ip
		}
	}
	fmt.Println("my ip", int_localIP, "master ip", highest_ip)
	if int_localIP == highest_ip {
		fmt.Println("I am the master")
		go bool_to_master_chan(true, is_master_chan)
	} else {
		fmt.Println("I am a slave")
		go bool_to_master_chan(false, is_master_chan)
	}

}

func Network_handler(is_master_chan chan bool) {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	var newInfo peers.PeerUpdate

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf(localIP)

		peerUpdateCh := make(chan peers.PeerUpdate)
		peerTxEnable := make(chan bool)

		go peers.Transmitter(20243, id, peerTxEnable)
		go peers.Receiver(20243, peerUpdateCh)
		for {
			select {
			case newInfoCatcher := <-peerUpdateCh:
				newInfo = newInfoCatcher

				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", newInfo.Peers)
				fmt.Printf("  New:      %q\n", newInfo.New)
				fmt.Printf("  Lost:     %q\n", newInfo.Lost)
				Choose_master(is_master_chan, newInfo)
			}
		}

	}
}

func Network_setup(master bool) {
	//THIS IS NOW RUNNING AS SLAVE
	var receivePort, broadcastPort int

	master_sender := make(chan Master_msg)
	master_receiver := make(chan Slave_msg)

	slave_sender := make(chan Slave_msg)
	slave_receiver := make(chan Master_msg)
	if master {
		fmt.Println("Connecting as the master")
		receivePort = slavePort
		broadcastPort = masterPort

		go bcast.Transmitter(broadcastPort, master_sender)
		go bcast.Receiver(receivePort, master_receiver)

		var msg Master_msg
		msg.Address = "Master sending msg"
		for {
			master_sender <- msg
			fmt.Println("master just sent")
			time.Sleep(1 * time.Second)
			//<-master_sender
			fmt.Println("chan empty")
		}

	} else {
		fmt.Println("Connecting as a slave")
		receivePort = masterPort
		broadcastPort = slavePort

		go bcast.Transmitter(broadcastPort, slave_sender)
		go bcast.Receiver(receivePort, slave_receiver)
		for {
			select {
			case a := <-slave_receiver: //burde hete slave_receiver_chan
				fmt.Println("The slave received : ", a)
			}
		}
	}
}
