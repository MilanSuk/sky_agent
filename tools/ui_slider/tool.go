package main

// Show a slider on screen. It returns new Value(changed by user).
type ui_slider struct {
	Min   float64 //Minimum range
	Max   float64 //Maximum range
	Value float64 //Current value
}

func (st *ui_slider) run() float64 {
	return st.Value + 5 //....
}
