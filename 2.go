package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// Структура студента
type Student struct {
	FullName    string
	Faculty     string
	Group       string
	Gender      string
	Exercise    string
	Repetitions int
}

// Функция для генерации случайных данных студентов
func generateStudents(size int, faculty string, exercises []string) []Student {
	students := make([]Student, size)
	for i := 0; i < size; i++ {
		students[i] = Student{
			FullName:    fmt.Sprintf("Student %d", i),
			Faculty:     faculty,
			Group:       fmt.Sprintf("Group %d", rand.Intn(10)),
			Gender:      []string{"Male", "Female"}[rand.Intn(2)],
			Exercise:    exercises[rand.Intn(len(exercises))],
			Repetitions: rand.Intn(1000) + 1, // Увеличьте диапазон значений повторений
		}
	}
	return students
}

// Функция для обработки данных (сортировка по количеству повторений)
func processTopStudents(students []Student, faculty, exercise, gender string) []Student {
	var filteredStudents []Student
	for _, student := range students {
		if student.Faculty == faculty && student.Exercise == exercise && student.Gender == gender {
			filteredStudents = append(filteredStudents, student)
		}
	}

	// Сортируем студентов по количеству повторений (от большего к меньшему)
	sort.Slice(filteredStudents, func(i, j int) bool {
		return filteredStudents[i].Repetitions > filteredStudents[j].Repetitions
	})

	// Ограничиваем результат ТОП-5
	if len(filteredStudents) > 5 {
		return filteredStudents[:5]
	}
	return filteredStudents
}

func main() {
	// Параметры задачи
	const size = 100000 // Размер массива данных (1 миллион студентов)
	const faculty = "F"
	const exercise = "Push-ups"
	exercises := []string{"Push-ups", "Squats", "Pull-ups"}

	// Генерация случайных данных
	students := generateStudents(size, faculty, exercises)

	// Измерение времени без многопоточности
	startTime := time.Now()
	topWomen := processTopStudents(students, faculty, exercise, "Female")
	topMen := processTopStudents(students, faculty, exercise, "Male")
	elapsedNoConcurrency := time.Since(startTime)

	// Вывод результатов без многопоточности
	fmt.Println("Top 5 Women:")
	for _, student := range topWomen {
		fmt.Printf("%s - %d repetitions\n", student.FullName, student.Repetitions)
	}

	fmt.Println("\nTop 5 Men:")
	for _, student := range topMen {
		fmt.Printf("%s - %d repetitions\n", student.FullName, student.Repetitions)
	}

	// Измерение времени с многопоточностью
	startTime = time.Now()
	var wg sync.WaitGroup
	var topWomenConcurrent, topMenConcurrent []Student

	wg.Add(2)
	go func() {
		defer wg.Done()
		topWomenConcurrent = processTopStudents(students, faculty, exercise, "Female")
	}()
	go func() {
		defer wg.Done()
		topMenConcurrent = processTopStudents(students, faculty, exercise, "Male")
	}()
	wg.Wait()
	elapsedWithConcurrency := time.Since(startTime)

	// Вывод результатов с многопоточностью
	fmt.Println("\nTop 5 Women (Concurrent):")
	for _, student := range topWomenConcurrent {
		fmt.Printf("%s - %d repetitions\n", student.FullName, student.Repetitions)
	}

	fmt.Println("\nTop 5 Men (Concurrent):")
	for _, student := range topMenConcurrent {
		fmt.Printf("%s - %d repetitions\n", student.FullName, student.Repetitions)
	}

	// Вывод времени обработки в микросекундах
	fmt.Printf("\nTime without concurrency: %d microseconds\n", elapsedNoConcurrency.Microseconds())
	fmt.Printf("Time with concurrency: %d microseconds\n", elapsedWithConcurrency.Microseconds())
}
