package fsm

import (
	"driver"
	"fmt"
	"global"
	"os"
	"queue"
	"time"
)

// TODO:
// -- Legge inn en go func med timer som sjekker om man er stuck
// -- Event idle reagerer ikke på updated orders, så viktig at heisene pusher new_order_chan ikke bare ved knappetrykk, men også hvis master assigner en ordre til dem
// -- Event idle sjekker nå external list. MÅ sjekke global list. Skal kun gjøre ting når de er blitt assigna til deg i global list
// -- Sjekker ikke om man skal stoppe når man kjører forbi etasjer

// Elevator states
const (
	Idle int = iota
	Moving
	Door_open
	Stuck
)

var Elev_state int

var current_order queue.Order

func State_handler(new_order_bool_chan chan bool, updated_order_bool_chan chan bool, update_order_chan chan queue.Order) {
	fmt.Println("Running: State handler")
	elev_state := Idle

	for {
		switch elev_state {
		case Idle:
			event_idle(new_order_bool_chan)
			Elev_state = Moving
			elev_state = Moving
		case Moving:
			event_moving()
			Elev_state = Door_open
			elev_state = Door_open
		case Door_open:
			event_door_open(update_order_chan)
			Elev_state = Idle
			elev_state = Idle
			time.Sleep(1*time.Second)
		case Stuck:
			os.Exit(0)

		}
	}
}

func event_idle(new_order_bool_chan chan bool) {
	fmt.Println(" I am in state Idle")
	is_order := false
	fmt.Println("Current order is before", current_order)

	for i := 0; i < global.NUM_INTERNAL_ORDERS; i++ {
		if queue.Internal_order_list[i].Order_state != queue.Inactive {
			if queue.Internal_order_list[i].Order_state != queue.Finished {
				current_order = queue.Internal_order_list[i]
				is_order = true
			}
		}

	}
	for i := 0; i < global.NUM_GLOBAL_ORDERS; i++ {
		if queue.External_order_list[i].Order_state != queue.Inactive {
			if queue.External_order_list[i].Order_state != queue.Finished {
				current_order = queue.External_order_list[i]
				is_order = true
			}
		}
	}
	fmt.Println("Current order is after", current_order)

	if is_order == false {
		select {
		case  <-new_order_bool_chan:
			//fmt.Println("Got new order bool ", catch_new_order_bool, " in Idle.")
			//fmt.Println("Now checking for orders that needs to be done inside the select case in event_idle")
			fmt.Println("Current order is", current_order)
			for i := 0; i < global.NUM_INTERNAL_ORDERS; i++ {
				if queue.Internal_order_list[i].Order_state != queue.Inactive {
					if queue.Internal_order_list[i].Order_state != queue.Finished {
						current_order = queue.Internal_order_list[i]
					}

				}
				for i := 0; i < global.NUM_GLOBAL_ORDERS; i++ {
					if queue.External_order_list[i].Order_state != queue.Inactive {
						if queue.External_order_list[i].Order_state != queue.Finished {
							current_order = queue.External_order_list[i]
						}
					}
				}
				fmt.Println("Current order is after", current_order)
			}
		}
	}	
}

func event_moving() {
	elevator_to_floor(current_order.Floor)
}

func event_door_open(update_order_chan chan queue.Order) {
	fmt.Println("Running event: Door open.")

	// Open door
	driver.Open_door()
	fmt.Println("Door opened.")
	driver.Set_button_lamp(current_order.Button, current_order.Floor, global.OFF) //-- can be moved to before open door
	fmt.Println("Door open lamp set on.")

	// Set order state to finished
	current_order.Order_state = queue.Finished
	fmt.Println("Current order state set to finished.")
	// ---- hmhmhmhmmh
	go queue.Order_to_update_order_chan(current_order, update_order_chan)
	fmt.Println("Order sent on updated order chan.")
}

func elevator_to_floor(floor global.Floor_t) {
	// Check if the elevator is between two floors
	timer := time.NewTimer(3 * time.Second)
	timeout := false
	go func() {
		<-timer.C
		timeout = true
	}()
	for driver.Get_floor_sensor_signal() == -1 {
		if !timeout {
			driver.Set_motor_direction(global.DIR_UP)
		} else if timeout {
			driver.Set_motor_direction(global.DIR_DOWN)
		}
	}

	// Go to desired floor
	current_floor_int := driver.Get_floor_sensor_signal()
	current_floor := driver.Floor_int_to_floor_t(current_floor_int)
	floor_int := driver.Floor_t_to_floor_int(floor)
	fmt.Println(current_floor_int, floor_int, current_floor)

	if current_floor_int < floor_int {
		fmt.Println("Going up.")
		driver.Set_motor_direction(global.DIR_UP)

		for driver.Get_floor_sensor_signal() != floor_int {
			current_floor = driver.Floor_int_to_floor_t(driver.Get_floor_sensor_signal())

			// When arriving at any floor, check for order
			if driver.Get_floor_sensor_signal() != -1 {
				driver.Set_floor_indicator_lamp(floor)
				//pick_up_order_on_the_way(current_floor, order_list, updated_order_chan, current_order)
				time.Sleep(10 * time.Millisecond)
			}
		}

	} else if current_floor_int > floor_int {
		fmt.Println("Going down.")
		driver.Set_motor_direction(global.DIR_DOWN)

		for driver.Get_floor_sensor_signal() != floor_int {
			current_floor = driver.Floor_int_to_floor_t(driver.Get_floor_sensor_signal())

			// When we arrive at any floor, check for order
			if driver.Get_floor_sensor_signal() != -1 {
				driver.Set_floor_indicator_lamp(floor)
				//pick_up_order_on_the_way(current_floor, order_list, updated_order_chan, current_order)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	// Stop when at desired floor
	driver.Set_motor_direction(global.DIR_STOP)
}
