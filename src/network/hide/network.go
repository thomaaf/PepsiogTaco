package network

import (
	"fmt"
	//"global"
	"flag"
	"queue"
	//"network/bcast"
	"network/localip"
	"network/peers"
	//"time"
)

// -- kan bruke json marshal greier for å pakke meldingen, og unpakke den
// -- sender det da som bytes, må være en public struct (stor forbokstav)

type Master_msg struct {
	Master_order_list [6]queue.Order
}

type Slave_msg struct {
	Slave_order_list [10]queue.Order
	Slave_info       queue.Elev_info
}

func Network_info() {
	fmt.Print("Running: Network info. ")
	//Kan brukes til å vite om masteren har falt ut (vet at det alltid er den med høyest IP). Hvis lengden av peers er lik 0
	//er man alene på nettverket.
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf(localIP)
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)

	//helloTx := make(chan string)
	//helloRx := make(chan string)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	//go bcast.Transmitter(20244, helloTx)
	//go bcast.Receiver(20244, helloRx)
	go peers.Transmitter(20243, id, peerTxEnable)
	go peers.Receiver(20243, peerUpdateCh)

	for {
		select {
		case newInfo := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", newInfo.Peers)
			fmt.Printf("  New:      %q\n", newInfo.New)
			fmt.Printf("  Lost:     %q\n", newInfo.Lost)
		}
	}

}

/*
func Test_network() {

	//Make channels for sending and receiving HelloMsg
	master_sender := make(chan Master_msg)
	master_receiver := make(chan Slave_msg)
	slave_sender := make(chan Slave_msg)
	slave_receiver := make(chan Master_msg)

	//Sier hvilken socket som skal gjøre hva
	go bcast.Transmitter(30000, master_sender)
	go bcast.Receiver(30000, master_receiver)
	go bcast.Transmitter(30000, slave_sender)
	go bcast.Receiver(30000, slave_receiver)

	//FIKSER SLAVE_INFO heeer ::::: //
	var slave_info queue.Elev_info
	slave_info.Elev_ip, _ = localip.LocalIP()
	slave_info.Elev_state = 0
	slave_info.Elev_last_floor = global.FLOOR_2
	slave_info.Elev_dir = global.DIR_UP

	go func() {
		master_message := Master_msg{queue.Global_order_list}
		slave_message := Slave_msg{my_order_list, slave_info}
		for {
			master_sender <- master_message
			slave_sender <- slave_message
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		select {
		case master := <-master_receiver:
			fmt.Println("Master receiving: ", master)
			time.Sleep(1 * time.Second)
		case slave := <-slave_receiver:
			fmt.Println("Slave receiving: ", slave)
			time.Sleep(1 * time.Second)
		}
	}
}*/

/*

// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf(localIP)
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)

	helloTx := make(chan string)
	helloRx := make(chan string)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(20244, helloTx)
	go bcast.Receiver(20244, helloRx)
	go peers.Transmitter(20243, id, peerTxEnable)
	go peers.Receiver(20243, peerUpdateCh)
	go func() {
		for {
			helloTx <- "HELLOOO"
			time.Sleep(1 * time.Second)
		}

	}()
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		case hellomessage := <-helloRx:
			fmt.Println("I received a message: ", hellomessage)
		}
	}


*/
