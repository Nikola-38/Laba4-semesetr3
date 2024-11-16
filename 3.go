package main

import (
	"fmt"
	"sync"
	"time"
)

const numPhilosophers = 5

// Структура для вилки, которая будет блокироваться при использовании
type Fork struct {
	mu sync.Mutex
}

// Структура философа
type Philosopher struct {
	id        int
	leftFork  *Fork
	rightFork *Fork
}

func (p *Philosopher) dine(wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 3; i++ { // философ ест 3 раза
		p.think()
		p.eat()
	}
}

func (p *Philosopher) think() {
	fmt.Printf("Философ %d думает\n", p.id)
	time.Sleep(time.Second) // время на размышления
}

func (p *Philosopher) eat() {
	// Блокируем вилки
	p.leftFork.mu.Lock()
	p.rightFork.mu.Lock()

	// Философ ест
	fmt.Printf("Философ %d ест спагетти\n", p.id)
	time.Sleep(time.Second * 2) // время на еду

	// Освобождаем вилки
	p.rightFork.mu.Unlock()
	p.leftFork.mu.Unlock()

	// Сообщаем, что философ закончил есть
	fmt.Printf("Философ %d закончил есть\n", p.id)
}

func main() {
	// Инициализация вилок
	forks := make([]*Fork, numPhilosophers)
	for i := range forks {
		forks[i] = &Fork{}
	}

	// Инициализация философов
	philosophers := make([]*Philosopher, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		philosophers[i] = &Philosopher{
			id:        i,
			leftFork:  forks[i],
			rightFork: forks[(i+1)%numPhilosophers], // правой вилкой является следующая в круге
		}
	}

	// Синхронизация с помощью WaitGroup
	var wg sync.WaitGroup
	for _, philosopher := range philosophers {
		wg.Add(1)
		go philosopher.dine(&wg)
	}

	// Ожидание завершения всех философов
	wg.Wait()
	fmt.Println("Все философы поели и завершили свою трапезу.")
}
