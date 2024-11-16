package main

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
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

// Функция для генерации случайного ASCII символа
func randomASCII() byte {
	return byte(rand.Intn(95) + 32) // Генерация символа с кодом от 32 до 126 (печатные символы)
}

func runWithMutex(n int) string {
	var result string
	mu.Lock()
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	mu.Unlock()
	return result
}

func runWithSemaphore(n int) string {
	var result string
	sem <- struct{}{} // Захват семафора
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	<-sem // Освобождение семафора
	return result
}

func runWithSemaphoreSlim(n int) string {
	var result string
	semaphoreSlim.Mutex.Lock()
	semaphoreSlim.sem <- struct{}{}
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	<-semaphoreSlim.sem
	semaphoreSlim.Mutex.Unlock()
	return result
}

func runWithSpinLock(n int) string {
	var result string
	spinLock.Lock()
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	spinLock.Unlock()
	return result
}

func runWithSpinWait(n int) string {
	var result string
	for !atomic.CompareAndSwapInt32(&spinWaitLock, 0, 1) {
	}
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	atomic.StoreInt32(&spinWaitLock, 0)
	return result
}

func runWithMonitor(m *Monitor, n int) string {
	var result string
	m.AccessResource()
	for i := 0; i < n; i++ {
		result += string(randomASCII())
	}
	m.ReleaseResource()
	return result
}

func BenchmarkMutex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithMutex(100)
	}
}

func BenchmarkSemaphore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithSemaphore(100)
	}
}

func BenchmarkSemaphoreSlim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithSemaphoreSlim(100)
	}
}

func BenchmarkSpinLock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithSpinLock(100)
	}
}

func BenchmarkSpinWait(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runWithSpinWait(100)
	}
}

func BenchmarkMonitor(b *testing.B) {
	monitor := NewMonitor()
	for i := 0; i < b.N; i++ {
		runWithMonitor(monitor, 100)
	}
}
