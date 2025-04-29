# TC02 - Simulador de Línea de Ensamblaje con IPC y Algoritmos de Scheduling

Este proyecto implementa un simulador de una línea de ensamblaje de productos en la que se utilizan mecanismos de comunicación entre procesos (IPC), junto con técnicas de sincronización (mutexes) para garantizar que sólo se procese un producto a la vez en cada estación. Se implementan dos algoritmos de scheduling: FCFS (First Come, First Serve) y Round Robin, siendo este último configurable en cuanto a su quantum.

---
### Miembros del grupo:
- Priscilla Jimenez Salgado
- Katerine Guzmán Flores
- Esteban Solano Araya
---

## 1. Instrucciones de Ejecución

### Requisitos
- **Go**: Asegúrate de tener instalado [Go](https://golang.org/dl/) (por ejemplo, la versión 1.24.2 o superior).

### Pasos para Ejecutar

1. **Clonar o Descargar el Repositorio:**
   - Coloca el código fuente en una carpeta de tu equipo.

2. **Abrir el Proyecto en la Terminal o en VSCode:**
   - Si usas VSCode, abre la carpeta del proyecto y ejecuta `go mod tidy` en la terminal integrada.

3. **Ejecutar el Programa:**
   - En la terminal, ubica la carpeta del proyecto y ejecuta:
     ```
     go run main.go
     ```
   - Se presentará un menú interactivo con dos opciones:
     - **1. FCFS**: Ejecuta el pipeline de estaciones en modo FCFS.
     - **2. Round Robin**: Ejecuta el pipeline en modo Round Robin.  
       *En este modo se te solicitará ingresar el quantum en segundos. Si el valor ingresado es inválido, se usará el quantum por defecto (2 segundos).*
     - **3. Salir**

4. **Visualizar los Resultados:**
   - Durante la ejecución, se mostrará el procesamiento de cada producto en cada estación, junto con mensajes de inicio y fin.
   - Al finalizar, se presentará un resumen de métricas que incluye:
     - Turnaround y tiempo de espera de cada producto.
     - Promedio de turnaround y de tiempo de espera.
     - Orden final de procesamiento.

5. **Repetir la Simulación (Opcional):**
   - Luego de terminar la ejecución, se te preguntará si deseas ejecutar otra simulación. Puedes repetir el proceso o seleccionar salir.

---

## 2. Descripción de la Solución

La solución se compone de los siguientes componentes:

- **Generación de Productos:**  
  La función `generarProductos` crea al menos 10 productos con un identificador único y un timestamp de llegada. Los productos se envían a un canal (cola de entrada) para ser procesados.

- **Estaciones de Ensamblaje:**  
  Se implementan tres estaciones que simulan una línea de ensamblaje (por ejemplo, Corte, Ensamblaje y Empaque).  
  - Cada estación se ejecuta como una goroutine y se comunica con la siguiente a través de canales.  
  - Se utiliza el tipo `EstacionIPC`, que incorpora un mutex para garantizar que la estación procese un solo producto a la vez.

- **Algoritmos de Scheduling:**  
  Se implementan dos modos de procesamiento:
  - **FCFS (First Come, First Serve):**  
    Los productos se procesan en el orden en que llegan (FIFO).  
    La función `TrabajarFCFS` utiliza la exclusión mutua (mutex) para proteger el procesamiento.
  - **Round Robin:**  
    Cada producto se procesa durante un tiempo fijo (quantum configurable). Si el producto no termina, se actualiza su tiempo restante y se reprograma para continuar en otra ronda.  
    La función `TrabajarRR` utiliza un mutex para proteger el acceso a la sección crítica durante cada ciclo.

- **Registro de Métricas:**  
  Se registran los tiempos de entrada y salida de cada producto en cada estación, se calcula el turnaround (tiempo total desde la llegada hasta la finalización) y el tiempo de espera.  
  Al finalizar, se muestra un resumen con los promedios y el orden final de procesamiento.

- **Menú Interactivo:**  
  La aplicación muestra un menú que permite seleccionar entre los algoritmos FCFS y Round Robin, con la opción de ingresar el quantum en el caso de Round Robin. Se permite ejecutar múltiples simulaciones sin reiniciar el programa.

---

## 3. Justificación Técnica

- **Uso de Go para Concurrencia e IPC:**  
  - Se aprovechan las goroutines para ejecutar de forma concurrente la generación de productos y el procesamiento en cada estación.
  - Los canales permiten la comunicación entre estaciones de forma segura y natural (FIFO), eliminando la necesidad de gestionar explícitamente el orden de llegada.

- **Sincronización con Mutexes:**  
  - Aunque los canales ya proporcionan una exclusión implícita, se incorporan mutexes en los algoritmos FCFS y Round Robin (en `EstacionIPC`) para cumplir con el criterio evaluativo de utilizar mecanismos de sincronización (exclusión mutua) explícitos.
  - Esto asegura que, aun en un entorno de alta concurrencia, cada estación procese un producto a la vez, evitando condiciones de carrera e interbloqueos.

- **Implementación de Algoritmos de Scheduling:**  
  - **FCFS:** Es sencillo de implementar, aprovechando la semántica FIFO de los canales.
  - **Round Robin:** Se implementa para permitir preempción mediante un quantum configurable. Esto permite que los productos con tiempos de procesamiento largos no bloqueen la línea de ensamblaje, mejorando la equidad en el uso de la estación.

- **Registro de Métricas:**  
  - Se utiliza un sistema de métricas para registrar los tiempos críticos del proceso y se presenta un resumen que permite evaluar el rendimiento del sistema.

---

## 4. Comparación entre Algoritmos FCFS y Round Robin

### FCFS (First Come, First Serve)
- **Ventajas:**
  - **Simplicidad:** Aprovecha directamente la naturaleza FIFO de los canales de Go.  
  - **Procesamiento Continuo:** Cada producto se procesa de principio a fin sin interrupciones.
- **Desventajas:**
  - **No Preemptivo:** Si un producto requiere un tiempo de procesamiento muy largo, puede causar demoras para los siguientes productos.  
  - **Inequidad:** No permite ajustar el reparto del tiempo de procesamiento si existen grandes diferencias en la complejidad o el tiempo requerido por cada producto.

### Round Robin
- **Ventajas:**
  - **Preempción:** Utiliza un quantum configurable que permite interrumpir el procesamiento y reprogramar productos que no hayan finalizado, lo que evita que un único producto monopolice la estación.
  - **Equidad:** Mayor equidad en la asignación del tiempo de procesamiento para productos con tiempos variables.
- **Desventajas:**
  - **Overhead Adicional:** La interrupción, registro y reprogramación de productos introduce un overhead extra en el procesamiento.
  - **Complejidad:** La implementación es un poco más compleja debido a la necesidad de administrar el tiempo restante y la reinserción en la cola.

---

Documentacion: https://estudianteccr-my.sharepoint.com/:w:/g/personal/guzka97_estudiantec_cr/EVQ8iqVdVVJAnDr0vhti0ZwB7d6KJgvjnXAhEineRavtdQ?e=42Ys8m 