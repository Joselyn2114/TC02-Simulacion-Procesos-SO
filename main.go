package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// Tipos y Funciones para Métricas
////////////////////////////////////////////////////////////////////////////////

// StationRecord almacena el instante de entrada y salida de un producto en una estación.
type StationRecord struct {
	StationID int
	EntryTime time.Time
	ExitTime  time.Time
}

// ProductMetric almacena las métricas de un producto (ID, llegada y registros por estación).
type ProductMetric struct {
	ProductID int
	Arrival   time.Time
	Records   []StationRecord
}

// LastExit devuelve el último instante en que el producto salió de alguna estación.
func (m *ProductMetric) LastExit() time.Time {
	var last time.Time
	for _, r := range m.Records {
		if r.ExitTime.After(last) {
			last = r.ExitTime
		}
	}
	return last
}

var (
	metricsMutex sync.Mutex
	metricsMap   = make(map[int]*ProductMetric)
)

func initProductMetric(p Producto) {
	metricsMutex.Lock()
	metricsMap[p.ID] = &ProductMetric{
		ProductID: p.ID,
		Arrival:   p.TiempoLlegada,
		Records:   []StationRecord{},
	}
	metricsMutex.Unlock()
}

func addStationRecord(productID int, stationID int, entry, exit time.Time) {
	metricsMutex.Lock()
	if m, ok := metricsMap[productID]; ok {
		m.Records = append(m.Records, StationRecord{
			StationID: stationID,
			EntryTime: entry,
			ExitTime:  exit,
		})
	}
	metricsMutex.Unlock()
}

func printMetricsSummary() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	var totalTurnaround, totalWaiting time.Duration
	count := len(metricsMap)
	fmt.Println("\n================ Resumen de Métricas ================")
	for _, m := range metricsMap {
		var processingTime time.Duration
		var lastExit time.Time
		for _, r := range m.Records {
			processingTime += r.ExitTime.Sub(r.EntryTime)
			if r.ExitTime.After(lastExit) {
				lastExit = r.ExitTime
			}
		}
		turnaround := lastExit.Sub(m.Arrival)
		waiting := turnaround - processingTime
		totalTurnaround += turnaround
		totalWaiting += waiting
		fmt.Printf("Producto #%d: Turnaround = %v, Tiempo de espera = %v\n",
			m.ProductID, turnaround, waiting)
	}
	avgTurnaround := time.Duration(0)
	avgWaiting := time.Duration(0)
	if count > 0 {
		avgTurnaround = totalTurnaround / time.Duration(count)
		avgWaiting = totalWaiting / time.Duration(count)
	}
	fmt.Printf("\nPromedio de Turnaround: %v\nPromedio de Tiempo de Espera: %v\n", avgTurnaround, avgWaiting)
	// Orden final de procesamiento (por último tiempo de salida)
	order := make([]*ProductMetric, 0, count)
	for _, m := range metricsMap {
		order = append(order, m)
	}
	sort.Slice(order, func(i, j int) bool {
		return order[i].LastExit().Before(order[j].LastExit())
	})
	fmt.Println("Orden final de procesamiento:")
	for _, m := range order {
		fmt.Printf("Producto #%d, finalizó a las %s\n", m.ProductID, m.LastExit().Format("15:04:05"))
	}
	fmt.Println("======================================================")
}

////////////////////////////////////////////////////////////////////////////////
// Tipos y Funciones Comunes (Tareas 3 y 4)
////////////////////////////////////////////////////////////////////////////////

// Producto representa un producto con ID único, tiempo de llegada y (para RR) tiempo restante.
type Producto struct {
	ID            int
	TiempoLlegada time.Time
	RemainingTime time.Duration // Se utiliza en Round Robin (inicialmente 0)
}

// generarProductos simula la generación de productos con retraso aleatorio y los envía a un canal.
// Cumple con la Tarea 3.
func generarProductos(numProductos int, out chan<- Producto) {
	for i := 1; i <= numProductos; i++ {
		delay := time.Duration(rand.Intn(3)+1) * time.Second
		time.Sleep(delay)
		producto := Producto{
			ID:            i,
			TiempoLlegada: time.Now(),
			RemainingTime: 0,
		}
		fmt.Printf("[%s] Producto #%d generado y en cola\n",
			producto.TiempoLlegada.Format("15:04:05"), producto.ID)
		initProductMetric(producto)
		out <- producto
	}
	close(out)
}

// Estacion utiliza un mutex para garantizar que se procese solo un producto a la vez.
// Cumple con la Tarea 4.
type Estacion struct {
	ID    int
	Mutex sync.Mutex
}

// ProcesarProducto procesa el producto usando exclusión mutua (Tareas 3 y 4 originales).
func (e *Estacion) ProcesarProducto(p Producto, wg *sync.WaitGroup) {
	defer wg.Done()
	e.Mutex.Lock()
	entry := time.Now()
	fmt.Printf("[%s] Producto #%d (llegó: %s) entrando en Estación #%d [ORIGINAL]\n",
		entry.Format("15:04:05"), p.ID, p.TiempoLlegada.Format("15:04:05"), e.ID)
	// Se simula un tiempo fijo de procesamiento (10 segundos)
	time.Sleep(10 * time.Second)
	exit := time.Now()
	fmt.Printf("[%s] Producto #%d ha sido procesado en Estación #%d [ORIGINAL]\n",
		exit.Format("15:04:05"), p.ID, e.ID)
	e.Mutex.Unlock()
	addStationRecord(p.ID, e.ID, entry, exit)
}

////////////////////////////////////////////////////////////////////////////////
// Implementación de Comunicación IPC y Scheduling (Tareas 1 y 2)
////////////////////////////////////////////////////////////////////////////////

// EstacionIPC representa una estación de trabajo comunicada mediante canales.
type EstacionIPC struct {
	ID       int
	Duracion time.Duration  // Tiempo total de procesamiento para la estación.
	In       <-chan Producto // Canal de entrada.
	Out      chan<- Producto // Canal de salida; nil en la última estación.
}

// TrabajarFCFS procesa productos en modo FCFS (FIFO) sin interrupción.
func (e *EstacionIPC) TrabajarFCFS(wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range e.In {
		entry := time.Now()
		fmt.Printf("[%s] EstaciónIPC #%d (FCFS): Inicia procesamiento del Producto #%d (llegó: %s)\n",
			entry.Format("15:04:05"), e.ID, p.ID, p.TiempoLlegada.Format("15:04:05"))
		time.Sleep(e.Duracion)
		exit := time.Now()
		fmt.Printf("[%s] EstaciónIPC #%d (FCFS): Finaliza procesamiento del Producto #%d\n",
			exit.Format("15:04:05"), e.ID, p.ID)
		addStationRecord(p.ID, e.ID, entry, exit)
		if e.Out != nil {
			e.Out <- p
		} else {
			fmt.Printf("[%s] Producto #%d completó la línea de ensamblaje (FCFS).\n",
				exit.Format("15:04:05"), p.ID)
		}
	}
	if e.Out != nil {
		close(e.Out)
	}
}

// TrabajarRR procesa productos en modo Round Robin con un quantum configurable.
// Si un producto no finaliza en el quantum, se reinserta en la cola para continuar su procesamiento.
func (e *EstacionIPC) TrabajarRR(wg *sync.WaitGroup, quantum time.Duration, expected int) {
	defer wg.Done()
	var rrQueue []Producto
	finished := 0

	drainInput := func() {
		for {
			select {
			case p, ok := <-e.In:
				if !ok {
					return
				}
				rrQueue = append(rrQueue, p)
			default:
				return
			}
		}
	}

	// Espera inicial para obtener el primer producto.
	p, ok := <-e.In
	if ok {
		rrQueue = append(rrQueue, p)
	} else {
		if e.Out != nil {
			close(e.Out)
		}
		return
	}

	for {
		drainInput()
		if len(rrQueue) == 0 {
			if finished >= expected {
				break
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		p = rrQueue[0]
		rrQueue = rrQueue[1:]
		if p.RemainingTime == 0 {
			p.RemainingTime = e.Duracion
		}
		entry := time.Now()
		if p.RemainingTime > quantum {
			fmt.Printf("[%s] EstaciónIPC #%d (RR): Procesando Producto #%d por quantum %v, tiempo restante %v\n",
				time.Now().Format("15:04:05"), e.ID, p.ID, quantum, p.RemainingTime)
			time.Sleep(quantum)
			p.RemainingTime -= quantum
			exit := time.Now()
			addStationRecord(p.ID, e.ID, entry, exit)
			rrQueue = append(rrQueue, p)
		} else {
			fmt.Printf("[%s] EstaciónIPC #%d (RR): Procesando Producto #%d por tiempo restante %v (finalizando)\n",
				time.Now().Format("15:04:05"), e.ID, p.ID, p.RemainingTime)
			time.Sleep(p.RemainingTime)
			exit := time.Now()
			addStationRecord(p.ID, e.ID, entry, exit)
			p.RemainingTime = 0
			finished++
			if e.Out != nil {
				e.Out <- p
			} else {
				fmt.Printf("[%s] Producto #%d completó la línea de ensamblaje (RR).\n",
					exit.Format("15:04:05"), p.ID)
			}
		}
		if finished >= expected && len(rrQueue) == 0 {
			break
		}
	}
	if e.Out != nil {
		close(e.Out)
	}
}

////////////////////////////////////////////////////////////////////////////////
// Función simulate: Ejecuta la simulación según el modo seleccionado.
////////////////////////////////////////////////////////////////////////////////

func simulate(schedulingMode string) {
	// Reinicia las métricas para cada simulación
	metricsMutex.Lock()
	metricsMap = make(map[int]*ProductMetric)
	metricsMutex.Unlock()

	numEstaciones := 3
	numProductos := 10

	if schedulingMode == "original" {
		// Modo ORIGINAL: Se usa la generación de productos y procesamiento con mutex.
		estaciones := make([]*Estacion, numEstaciones)
		for i := 0; i < numEstaciones; i++ {
			estaciones[i] = &Estacion{ID: i + 1}
		}
		productQueue := make(chan Producto, numProductos)
		var wg sync.WaitGroup
		go generarProductos(numProductos, productQueue)
		for producto := range productQueue {
			wg.Add(1)
			// Asigna de forma cíclica el producto a una estación.
			estacion := estaciones[(producto.ID-1)%numEstaciones]
			go estacion.ProcesarProducto(producto, &wg)
		}
		wg.Wait()
		fmt.Println("Todos los productos han sido procesados correctamente [ORIGINAL].")
	} else if schedulingMode == "fcfs" || schedulingMode == "rr" {
		// Modo IPC: Se crean canales para conectar las estaciones.
		canal1 := make(chan Producto, numProductos)
		canal2 := make(chan Producto, numProductos)
		canal3 := make(chan Producto, numProductos)
		var wg sync.WaitGroup
		wg.Add(3) // Tres estaciones

		estacion1 := EstacionIPC{
			ID:       1,
			Duracion: 3 * time.Second,
			In:       canal1,
			Out:      canal2,
		}
		estacion2 := EstacionIPC{
			ID:       2,
			Duracion: 4 * time.Second,
			In:       canal2,
			Out:      canal3,
		}
		estacion3 := EstacionIPC{
			ID:       3,
			Duracion: 5 * time.Second,
			In:       canal3,
			Out:      nil, // Última estación
		}

		if schedulingMode == "fcfs" {
			go estacion1.TrabajarFCFS(&wg)
			go estacion2.TrabajarFCFS(&wg)
			go estacion3.TrabajarFCFS(&wg)
		} else { // "rr"
			// Se solicita al usuario el quantum para RR
			fmt.Print("Ingrese el quantum (en segundos) para Round Robin: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error leyendo el quantum. Se usará el valor por defecto (2 segundos).")
				input = "2"
			}
			input = strings.TrimSpace(input)
			quantumSec, err := strconv.Atoi(input)
			if err != nil || quantumSec <= 0 {
				fmt.Println("Valor inválido. Se usará el valor por defecto (2 segundos).")
				quantumSec = 2
			}
			quantum := time.Duration(quantumSec) * time.Second

			go estacion1.TrabajarRR(&wg, quantum, numProductos)
			go estacion2.TrabajarRR(&wg, quantum, numProductos)
			go estacion3.TrabajarRR(&wg, quantum, numProductos)
		}
		go generarProductos(numProductos, canal1)
		wg.Wait()
		fmt.Printf("Todos los productos han sido procesados correctamente [%s].\n", schedulingMode)
	} else {
		fmt.Println("Modo de scheduling desconocido.")
		return
	}
	printMetricsSummary()
}

////////////////////////////////////////////////////////////////////////////////
// Función Main: Menú Interactivo
////////////////////////////////////////////////////////////////////////////////

func main() {
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n======================================")
		fmt.Println("Seleccione el modo de ejecución:")
		fmt.Println("1. Original (Sincronización con Mutexes)")
		fmt.Println("2. FCFS (Pipeline con IPC en modo FCFS)")
		fmt.Println("3. Round Robin (Pipeline con IPC en modo RR)")
		fmt.Println("4. Salir")
		fmt.Print("Ingrese opción (1-4): ")

		choiceStr, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error leyendo la entrada. Intente de nuevo.")
			continue
		}
		var schedulingMode string
		switch choiceStr[0] {
		case '1':
			schedulingMode = "original"
		case '2':
			schedulingMode = "fcfs"
		case '3':
			schedulingMode = "rr"
		case '4':
			fmt.Println("Saliendo del programa.")
			return
		default:
			fmt.Println("Opción inválida. Intente nuevamente.")
			continue
		}

		// Ejecuta la simulación con el modo seleccionado.
		simulate(schedulingMode)

		fmt.Print("¿Desea ejecutar otra simulación? (s/n): ")
		resp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error leyendo la respuesta. Saliendo.")
			return
		}
		if resp[0] != 's' && resp[0] != 'S' {
			fmt.Println("Saliendo del programa.")
			break
		}
	}
}
