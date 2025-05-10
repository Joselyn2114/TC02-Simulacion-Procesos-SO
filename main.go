package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Definimos una estación de trabajo
type Estacion struct {
	ID    int
	Mutex sync.Mutex // Mutex para evitar que múltiples productos se procesen a la vez
}

// Producto representa un producto con un identificador único y un tiempo de llegada.
type Producto struct {
	ID            int
	TiempoLlegada time.Time
}

// ProcesarProducto simula el procesamiento de un producto en la estación.
func (e *Estacion) ProcesarProducto(p Producto, wg *sync.WaitGroup) {
	defer wg.Done()

	e.Mutex.Lock() // Bloqueamos la estación para este producto
	horaInicio := time.Now().Format("15:04:05")
	fmt.Printf("[%s] Producto # %d (llegó: %s) entrando en Estación # %d\n",
		horaInicio, p.ID, p.TiempoLlegada.Format("15:04:05"), e.ID)

	// Simulamos el tiempo de procesamiento (por ejemplo, 10 segundos)
	time.Sleep(time.Second * 10)

	horaFin := time.Now().Format("15:04:05")
	fmt.Printf("[%s] Producto # %d ha sido procesado en Estación # %d\n",
		horaFin, p.ID, e.ID)
	e.Mutex.Unlock() // Liberamos la estación para otro producto
}

// generarProductos simula la generación de productos con tiempos de llegada y los coloca en una cola.
func generarProductos(numProductos int, productQueue chan<- Producto) {
	for i := 1; i <= numProductos; i++ {
		// Simulamos el tiempo de llegada con un retraso aleatorio entre 1 y 3 segundos.
		delay := time.Duration(rand.Intn(3)+1) * time.Second
		time.Sleep(delay)

		producto := Producto{
			ID:            i,
			TiempoLlegada: time.Now(),
		}

		fmt.Printf("[%s] Producto # %d generado y en cola\n",
			producto.TiempoLlegada.Format("15:04:05"), producto.ID)

		// Enviamos el producto a la cola
		productQueue <- producto
	}
	close(productQueue)
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Inicializamos la semilla para números aleatorios

	numEstaciones := 3
	numProductos := 5

	// Crear estaciones
	estaciones := make([]*Estacion, numEstaciones)
	for i := 0; i < numEstaciones; i++ {
		estaciones[i] = &Estacion{ID: i + 1}
	}

	// Canal que actuará como cola para almacenar productos
	productQueue := make(chan Producto, numProductos)

	var wg sync.WaitGroup //  WaitGroup para esperar a que se procesen todos los productos

	// Iniciamos la generación de productos en una goroutine
	go generarProductos(numProductos, productQueue)

	// Mientras existan productos en la cola, se los asigna a las estaciones (en este ejemplo, usando Round Robin)
	for producto := range productQueue {
		wg.Add(1)
		// Seleccionamos una estación en forma cíclica
		estacion := estaciones[(producto.ID-1)%numEstaciones]
		go estacion.ProcesarProducto(producto, &wg)
	}

	wg.Wait() // Esperamos a que todas las goroutines terminen
	fmt.Println("Todos los productos han sido procesados correctamente.")
}

