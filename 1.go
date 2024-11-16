package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Примитивы синхронизации
var mu sync.Mutex                // Mutex
var sem = make(chan struct{}, 3) // Semaphore с ограничением 3
var semaphoreSlim = struct {
	sync.Mutex
	sem chan struct{}
}{
	sem: make(chan struct{}, 1),
}
var wg sync.WaitGroup  // Для синхронизации в случае с барьером
var spinLock SpinLock  // SpinLock
var spinWaitLock int32 // SpinWait

// Структура для SpinLock
type SpinLock struct {
	lock int32
}

func (s *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&s.lock, 0, 1) {
	}
}

func (s *SpinLock) Unlock() {
	atomic.StoreInt32(&s.lock, 0)
}

// Структура для Monitor
type Monitor struct {
	mu       sync.Mutex
	cond     *sync.Cond
	resource int
}

func NewMonitor() *Monitor {
	m := &Monitor{}
	m.cond = sync.NewCond(&m.mu)
	return m
}

func (m *Monitor) AccessResource() {
	m.mu.Lock()
	for m.resource >= 3 { // Допустим, максимальное количество ресурса — 3
		m.cond.Wait() // Ожидаем, пока не будет доступен ресурс
	}
	m.resource++
	m.mu.Unlock()
}

func (m *Monitor) ReleaseResource() {
	m.mu.Lock()
	m.resource--
	m.cond.Signal() // Освобождаем один ресурс
	m.mu.Unlock()
}

// Структура для StopWatch
type StopWatch struct {
	start time.Time
}

func (s *StopWatch) Start() {
	s.start = time.Now()
}

func (s *StopWatch) Elapsed() time.Duration {
	return time.Since(s.start)
}

// Функция для генерации случайного ASCII символа
func randomASCII() byte {
	return byte(rand.Intn(95) + 32) // Генерация символа с кодом от 32 до 126 (печатные символы)
}

// Основная функция для выполнения задачи с разными примитивами
func runWithMutex(n int, wg *sync.WaitGroup) {
	defer wg.Done() // Уменьшаем счётчик WaitGroup по завершению горутины

	var result string
	mu.Lock()
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	mu.Unlock()
	fmt.Println("Result with Mutex:", result)
}

func runWithSemaphore(n int, wg *sync.WaitGroup) {
	defer wg.Done()

	var result string
	sem <- struct{}{} // Захват семафора
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	<-sem // Освобождение семафора
	fmt.Println("Result with Semaphore:", result)
}

func runWithSemaphoreSlim(n int, wg *sync.WaitGroup) {
	defer wg.Done()

	var result string
	semaphoreSlim.Mutex.Lock()
	semaphoreSlim.sem <- struct{}{}
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	<-semaphoreSlim.sem
	semaphoreSlim.Mutex.Unlock()
	fmt.Println("Result with SemaphoreSlim:", result)
}

func runWithSpinLock(n int, wg *sync.WaitGroup) {
	defer wg.Done()

	var result string
	spinLock.Lock()
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	spinLock.Unlock()
	fmt.Println("Result with SpinLock:", result)
}

func runWithSpinWait(n int, wg *sync.WaitGroup) {
	defer wg.Done()

	var result string
	for !atomic.CompareAndSwapInt32(&spinWaitLock, 0, 1) {
	}
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	atomic.StoreInt32(&spinWaitLock, 0)
	fmt.Println("Result with SpinWait:", result)
}

// Пример использования Barrier с sync.WaitGroup
func runWithBarrier(n int, barrier *sync.WaitGroup) {
	defer barrier.Done()

	var result string
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	fmt.Println("Result with Barrier:", result)
}

// Пример использования Monitor
func runWithMonitor(m *Monitor, n int, wg *sync.WaitGroup) {
	defer wg.Done()

	var result string
	m.AccessResource()
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	m.ReleaseResource()
	fmt.Println("Result with Monitor:", result)
}

func main() {
	// Задание для тестирования
	const numGoroutines = 5
	const numChars = 1

	// Инициализация Monitor
	monitor := NewMonitor()

	// Инициализация StopWatch
	sw := StopWatch{}

	// Замер общего времени выполнения программы
	sw.Start()

	// 1. Замер времени для Mutex
	startTime := time.Now()
	var wgMutex sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wgMutex.Add(1) // Увеличиваем счётчик WaitGroup
		go runWithMutex(numChars, &wgMutex)
	}
	wgMutex.Wait() // Ожидаем завершения всех горутин
	mutexTime := time.Since(startTime)
	fmt.Printf("Mutex execution took: %.2fns\n", float64(mutexTime.Nanoseconds()))

	// 2. Замер времени для Semaphore
	startTime = time.Now()
	var wgSemaphore sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wgSemaphore.Add(1)
		go runWithSemaphore(numChars, &wgSemaphore)
	}
	wgSemaphore.Wait()
	semaphoreTime := time.Since(startTime)
	fmt.Printf("Semaphore execution took: %.2fns\n", float64(semaphoreTime.Nanoseconds()))

	// 3. Замер времени для SemaphoreSlim
	startTime = time.Now()
	var wgSemaphoreSlim sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wgSemaphoreSlim.Add(1)
		go runWithSemaphoreSlim(numChars, &wgSemaphoreSlim)
	}
	wgSemaphoreSlim.Wait()
	semaphoreSlimTime := time.Since(startTime)
	fmt.Printf("SemaphoreSlim execution took: %.2fns\n", float64(semaphoreSlimTime.Nanoseconds()))

	// 4. Замер времени для SpinLock
	startTime = time.Now()
	var wgSpinLock sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wgSpinLock.Add(1)
		go runWithSpinLock(numChars, &wgSpinLock)
	}
	wgSpinLock.Wait()
	spinLockTime := time.Since(startTime)
	fmt.Printf("SpinLock execution took: %.2fns\n", float64(spinLockTime.Nanoseconds()))

	// 5. Замер времени для SpinWait
	startTime = time.Now()
	var wgSpinWait sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wgSpinWait.Add(1)
		go runWithSpinWait(numChars, &wgSpinWait)
	}
	wgSpinWait.Wait()
	spinWaitTime := time.Since(startTime)
	fmt.Printf("SpinWait execution took: %.2fns\n", float64(spinWaitTime.Nanoseconds()))

	// 6. Замер времени для Barrier (с использованием sync.WaitGroup)
	startTime = time.Now()
	var wgBarrier sync.WaitGroup
	wgBarrier.Add(numGoroutines) // Устанавливаем количество горутин, которые должны завершиться
	for i := 0; i < numGoroutines; i++ {
		go runWithBarrier(numChars, &wgBarrier)
	}
	wgBarrier.Wait() // Ждем завершения всех горутин
	barrierTime := time.Since(startTime)
	fmt.Printf("Barrier execution took: %.2fns\n", float64(barrierTime.Nanoseconds()))

	// 7. Замер времени для Monitor
	startTime = time.Now()
	var wgMonitor sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wgMonitor.Add(1)
		go runWithMonitor(monitor, numChars, &wgMonitor)
	}
	wgMonitor.Wait()
	monitorTime := time.Since(startTime)
	fmt.Printf("Monitor execution took: %.2fns\n", float64(monitorTime.Nanoseconds()))

	// Выводим общее время выполнения
	totalTime := time.Since(sw.start)
	fmt.Printf("Total execution time: %.2fns\n", float64(totalTime.Nanoseconds()))
}
