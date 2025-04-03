package main

import (
	"fmt"
	"sync"
	"time"
)

// Definimos una estación de trabajo
type Estacion struct {
	ID    int
	Mutex sync.Mutex // Mutex para evitar que múltiples productos se procesen a la vez
}

// Simula el procesamiento de un producto en una estación
func (e *Estacion) ProcesarProducto(idProducto int, wg *sync.WaitGroup) {
	defer wg.Done() // Marca esta tarea como terminada cuando salga

	e.Mutex.Lock() // Bloqueamos la estación para este producto
	fmt.Printf("📌 Producto %d entrando en Estación %d\n", idProducto, e.ID)
	time.Sleep(time.Second * 2) // Simulamos procesamiento
	fmt.Printf("✅ Producto %d ha sido procesado en Estación %d\n", idProducto, e.ID)
	e.Mutex.Unlock() // Liberamos la estación para otro producto
}

func main() {
	numEstaciones := 3
	numProductos := 5

	// Crear estaciones
	estaciones := make([]*Estacion, numEstaciones)
	for i := 0; i < numEstaciones; i++ {
		estaciones[i] = &Estacion{ID: i + 1}
	}

	var wg sync.WaitGroup // Controla la finalización de goroutines

	// Enviar productos a estaciones
	for i := 1; i <= numProductos; i++ {
		wg.Add(1) // Añadimos una tarea al WaitGroup
		go estaciones[i%numEstaciones].ProcesarProducto(i, &wg)
	}

	wg.Wait() // Esperamos a que todas las goroutines terminen
	fmt.Println("🎯 Todos los productos han sido procesados correctamente.")
}
