package main

type formIdSelect struct {
	id int64
	title string
}

type formIdSelectsMsg struct {
	items []formIdSelect
}
