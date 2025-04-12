# TC02-Simulacion-Procesos-SO

## Descripción de la Solución
Este proyecto simula una línea de ensamblaje de productos en la que:
- Se generan al menos 10 productos, cada uno con un identificador único y un timestamp simulado de llegada.
- La línea está compuesta por tres estaciones de trabajo (por ejemplo, Corte → Ensamblaje → Empaque).
- La comunicación entre estaciones se realiza mediante canales (pipes o colas compartidas) en el modo IPC.
- Se implementan dos algoritmos de scheduling:
  - **FCFS (First Come, First Serve):** Procesa los productos en el orden de llegada.
  - **Round Robin:** Utiliza un quantum configurable para procesar cada producto; si no finaliza, se reprograma para continuar su procesamiento.
- Se aplican técnicas de sincronización:
  - En el modo original se usa `sync.Mutex` para garantizar que solo se procese un producto a la vez.
  - En los modos IPC, la secuencia en la lectura de los canales asegura la exclusión mutua.
- Se registran métricas de rendimiento (tiempo de llegada, entrada y salida en cada estación, turnaround y tiempo de espera) y se muestra un resumen final.

## Justificación Técnica
- **Modelado de Concurrencia e IPC:**  
  El uso de goroutines y canales en Go permite una implementación natural de la comunicación entre estaciones, garantizando que cada producto se procese de manera secuencial en cada etapa.
- **Sincronización:**  
  La solución aprovecha tanto la exclusión mutua a través de `sync.Mutex` (en el modo original) como el comportamiento FIFO natural de los canales para evitar condiciones de carrera.
- **Algoritmos de Scheduling:**  
  Se implementa FCFS directamente con la semántica FIFO de los canales. El algoritmo Round Robin se logra interrumpiendo el procesamiento (por un quantum configurable) y reprogramando los productos incompletos, lo que previene la espera indefinida y permite manejar productos con tiempos de procesamiento variables.
- **Registro de Métricas:**  
  Se registra el tiempo de entrada y salida en cada estación, se calcula el turnaround y el tiempo de espera, proporcionando información valiosa sobre el rendimiento del sistema.

## Comparación entre Algoritmos
- **FCFS:**
  - **Ventajas:**  
    - Implementación sencilla aprovechando la naturaleza FIFO de los canales.
    - Procesamiento sin interrupciones, lo que puede ser adecuado para tiempos de procesamiento homogéneos.
  - **Desventajas:**  
    - No permite la preempción; productos con tiempos de procesamiento muy largos pueden afectar la fluidez de la línea.
  
- **Round Robin:**
  - **Ventajas:**  
    - Permite la preempción mediante un quantum configurable, lo que equilibra el tiempo de procesamiento entre productos.
    - Previene la espera indefinida (starvation) y mejora la equidad en el uso de recursos.
  - **Desventajas:**  
    - La administración del quantum y la reinserción de productos añade complejidad.
    - Puede incrementar el overhead debido a frecuentes interrupciones y reprogramaciones.

## Instrucciones de Ejecución
1. **Instalación de Go:**  
   Descarga e instala [Go](https://golang.org/dl/) (versión recomendada: 1.24.2 o superior).

2. **Descarga del Proyecto:**  
   Clona o descarga el repositorio del proyecto en tu equipo.

3. **Configuración en VSCode (opcional):**
   - Abre la carpeta del proyecto en Visual Studio Code.
   - Instala la extensión oficial de Go.
   - Abre la terminal integrada y ejecuta:
     ```
     go mod tidy
     ```

4. **Ejecución del Programa:**
   - En la terminal, ejecuta:
     ```
     go run main.go
     ```
   - Se mostrará un menú interactivo:
     - Selecciona **1** para el modo *Original* (procesamiento con mutex).
     - Selecciona **2** para el modo *FCFS* (pipeline IPC con procesamiento en orden FIFO).
     - Selecciona **3** para el modo *Round Robin* (pipeline IPC con quantum configurable).
     - Selecciona **4** para salir.
   - Tras la ejecución se mostrarán las métricas y se preguntará si deseas ejecutar otra simulación.

5. **Visualización de Resultados:**  
   Observa en la terminal la secuencia de procesamiento, así como el resumen de métricas (turnaround, tiempo de espera y orden final de procesamiento).

Documentacion: https://estudianteccr-my.sharepoint.com/:w:/g/personal/guzka97_estudiantec_cr/EVQ8iqVdVVJAnDr0vhti0ZwB7d6KJgvjnXAhEineRavtdQ?e=42Ys8m 