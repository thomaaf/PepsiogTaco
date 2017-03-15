package network

import (
	"flag"
	"fmt"
	"global"
	//"network/bcast"
	"network/bcast"
	"network/localip"
	"network/peers"
	"queue"
	"strconv"
	"time"
)

const (
	master_port = 20079
	slave_port  = 20179
)

var Local_ip int
var Is_master bool
var Num_elev_online int
var Lost_network bool

var Elevators_online [3]int

type Master_msg struct {
	Address     int
	Global_list [global.NUM_GLOBAL_ORDERS]queue.Order
}

type Slave_msg struct {
	Address       int
	Internal_list [global.NUM_INTERNAL_ORDERS]queue.Order
	External_list [global.NUM_GLOBAL_ORDERS]queue.Order
	Elevator_info queue.Elev_info
}

func Choose_master() {
	ip_addresses := Elevators_online
	Num_elev_online = 0
	highest_ip := 0
	for i := 0; i < 2; i++ {
		if ip_addresses[i] > highest_ip {
			highest_ip = ip_addresses[i]
		}
	}
	if Local_ip == highest_ip {
		fmt.Println("I am the master.")
		Is_master = true
	} else {
		fmt.Println("I am a slave.")
		Is_master = false
	}
	if highest_ip == 0 {
		fmt.Println("I have lost network")
		Is_master = false
		Lost_network = true
	}
}

func Network_handler() {
	var new_info peers.PeerUpdate

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	//localip, _ := localip.LocalIP()

	if id == "" {
		local_ip, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			local_ip = "Disconnected."
		}
		id = fmt.Sprintf(local_ip)

		peer_update_chan := make(chan peers.PeerUpdate)
		peer_transmit_enable_chan := make(chan bool)

		go peers.Transmitter(20243, id, peer_transmit_enable_chan)
		go peers.Receiver(20243, peer_update_chan)
		for {
			select {
			case new_info_catcher := <-peer_update_chan:
				Num_elev_online = 0
				new_info = new_info_catcher

				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", new_info.Peers)
				fmt.Printf("  New:      %q\n", new_info.New)
				fmt.Printf("  Lost:     %q\n", new_info.Lost)
				for i := 0; i < len(new_info.Peers); i++ {
					str_ip := new_info.Peers[i]
					int_ip, _ := strconv.Atoi(str_ip[12:])
					Elevators_online[i] = int_ip
					Num_elev_online = Num_elev_online + 1
				}
				if len(new_info.Peers) == 2 {
					Elevators_online[2] = -1
				}
				if len(new_info.Peers) == 1 {
					Elevators_online[1] = -1
					Elevators_online[2] = -1
				}
				if len(new_info.Peers) == 0 {
					Elevators_online[1] = -1
					Elevators_online[2] = -1
					Elevators_online[0] = -1

				}
				Local_ip, _ = strconv.Atoi(local_ip[12:])
				Choose_master()
			}
		}
	}
}

func Network_setup(new_order_bool_chan chan bool) {
	var receive_port, broadcast_port int

	master_sender := make(chan Master_msg)
	master_receiver := make(chan Slave_msg)
	slave_sender := make(chan Slave_msg)
	slave_receiver := make(chan Master_msg)

	if Is_master {
		fmt.Println("Connecting as the master.")
		receive_port = slave_port
		broadcast_port = master_port

		go bcast.Transmitter(broadcast_port, master_sender)
		go bcast.Receiver(receive_port, master_receiver)
		go master_transmit(master_sender)

		for {
			select {
			case catch_msg_from_slave := <-master_receiver:
				fmt.Println("Master received : ", catch_msg_from_slave)
				//queue.Master_msg_handler(catch_msg_from_slave)
			}
		}
	} else {
		fmt.Println("Connecting as a slave.")
		receive_port = master_port
		broadcast_port = slave_port

		go bcast.Transmitter(broadcast_port, slave_sender)
		go bcast.Receiver(receive_port, slave_receiver)
		go slave_transmit(slave_sender)

		for {
			select {
			case catch_msg_from_master := <-slave_receiver:
				fmt.Println("Slave received : ", catch_msg_from_master)
				//queue.Slave_msg_handler(catch_msg_from_master, new_order_bool_chan)
			}
		}
	}
}

func master_transmit(master_sender chan Master_msg) {
	var master_msg_to_send Master_msg
	for {
		master_msg_to_send.Address = Local_ip
		master_msg_to_send.Global_list = queue.Global_order_list
		master_sender <- master_msg_to_send
		fmt.Println("Master sent the global list: ", master_msg_to_send.Global_list)
		time.Sleep(1 * time.Second)
	}
}

func slave_transmit(slave_sender chan Slave_msg) {
	var slave_msg_to_send Slave_msg
	for {
		slave_msg_to_send.Address = Local_ip
		slave_msg_to_send.Internal_list = queue.Internal_order_list
		slave_msg_to_send.External_list = queue.External_order_list
		//slave_msg_to_send.Elevator_info = queue.Elev_info
		slave_sender <- slave_msg_to_send
		fmt.Println("Slave sent the lists: ")
		time.Sleep(1 * time.Second)
	}
}
