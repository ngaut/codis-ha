package main

type AliveChecker interface {
	CheckAlive() error
	Promote() error
}
