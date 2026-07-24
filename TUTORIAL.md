# Timescan: Tutorial Completo Passo-Passo

Benvenuto nel tutorial ufficiale di **Timescan**! Questa guida ti accompagnerà passo dopo passo nella costruzione di un'applicazione completa in Go che sfrutta tutte le funzionalità della libreria:

1. **Ingestione in Tempo Reale** con gestione della memoria Zero-Allocation.
2. **Motore di Pipeline** e Coordinamento dei flussi.
3. **Rilevamento di Anomalie** tramite algoritmi avanzati (`EWMA`, `Z-Score`, `MAD`).
4. **Compressione Vettoriale dei Pattern** (`PAAEncode`).
5. **Integrazione con Vector Database** (`Qdrant` / `Bbolt`).
6. **Decomposizione della Stagionalità** (`Trend`, `Seasonal`, `Residual`).
7. **Statistiche in Real-Time ed Offline** (Algoritmo di Welford, Mediana, IQR).

---

## Passo 1: Definire il Modello dei Dati

In Timescan ogni punto di misurazione è un `timeseries.DataPoint`:

```go
import "github.com/timescan/timescan/timeseries"

dp := timeseries.DataPoint{
    Timestamp: time.Now(),
    Value:     142.5,
    Tags:      map[string]string{"host": "server-01", "metric": "cpu_usage"},
}
```

---

## Passo 2: Inizializzare il Vector Store Adapter

Timescan permette di salvare le "forme" delle anomalie su un database vettoriale. Inizializziamo il driver per **Qdrant**:

```go
import "github.com/timescan/timescan/vector/driver/qdrant"

vecStore, err := qdrant.NewStore(qdrant.Config{
    Addr:       "localhost:6334",
    Collection: "incident_patterns",
})
```

---

## Passo 3: Configurare il Motore di Pipeline e Rilevamento

Creiamo la `pipeline.Engine` che gestirà il buffer scorrevole (`RingBufferWindow`) e collegherà il rilevatore di anomalie (`EWMADetector`):

```go
import (
    "github.com/timescan/timescan/anomaly"
    "github.com/timescan/timescan/pipeline"
)

detector := anomaly.NewEWMA(anomaly.EWMAConfig{
    Alpha:     0.15, // Peso dei dati recenti
    Threshold: 2.5,  // Soglia di allarme (in deviazioni standard)
})

engine := pipeline.NewEngine(pipeline.Config{
    WindowSize:  60,       // Mantiene gli ultimi 60 punti in memoria
    Detector:    detector,
    VectorStore: vecStore,
})
```

---

## Passo 4: Processare il Flusso in Real-Time (Zero-Alloc)

Ad ogni dato in arrivo chiamiamo `engine.Process(dp)`. Se viene trovata un'anomalia, l'engine restituisce i dettagli nel `Result`:

```go
result := engine.Process(dp)

if result.IsAnomaly {
    fmt.Printf("[ALERT] Anomalia Rilevata! Valore: %.2f (Atteso: %.2f, Score: %.2f)\n",
        dp.Value, result.AnomalyMeta.Expected, result.AnomalyMeta.Score)
}
```

---

## Passo 5: Convertire la Forma in Vettore (`PAAEncode`) e Cercare Similitudini

Quando viene rilevata un'anomalia, prendiamo la finestra di 60 punti temporali dal `result.WindowContext` e la comprimiamo in un vettore a 8 dimensioni tramite l'algoritmo **PAA** (Piecewise Aggregate Approximation):

```go
import "github.com/timescan/timescan/vector"

// Comprime 60 punti in 8 dimensioni
dimensions := 8
vectorEmbedding := vector.PAAEncode(result.WindowContext, dimensions)

// 1. Salva il pattern su Qdrant
_ = vecStore.Upsert(context.Background(), "incident-101", vectorEmbedding, map[string]any{
    "host": dp.Tags["host"],
})

// 2. Cerca nel database se questo tipo di guasto si è già verificato
matches, _ := vecStore.SearchNearest(context.Background(), vectorEmbedding, 3, nil)
```

---

## Passo 6: Decomposizione della Stagionalità (Offline)

Per analizzare una serie temporale complessa ed estrarre la stagionalità (es. cicli di traffico giornalieri/settimanali):

```go
import "github.com/timescan/timescan/decomposition"

period := 7 // Ciclo di 7 giorni
decomp := decomposition.DecomposeAdditive(historicalSeries, period)

fmt.Printf("Trend Generale: %.2f\n", decomp.Trend.Points[15].Value)
fmt.Printf("Effetto Stagionale: %.2f\n", decomp.Seasonal.Points[15].Value)
fmt.Printf("Rumore Residuo: %.2f\n", decomp.Residual.Points[15].Value)
```

---

## Passo 7: Calcolo Statistiche Online ed Offline

Timescan include strumenti per il calcolo di statistiche senza impattare la memoria:

```go
// 1. Statistiche Online in Tempo Reale (Welford)
welford := timeseries.NewWelford()
welford.Update(100.5)
fmt.Printf("Media Mobile: %.2f, DevDev: %.2f\n", welford.Mean(), welford.StdDev())

// 2. Statistiche Offline Robuste (Mediana e MAD)
values := []float64{10, 12, 11, 13, 1000} // 1000 è un outlier estremo
med := timeseries.Median(values)          // Restituisce 12.00 (non influenzato da 1000)
mad := timeseries.MAD(values)             // Deviazione Assoluta della Mediana
```

---

## Codice Completo ed Eseguibile

Tutti questi passaggi sono stati assemblati in un unico programma Go pronto per l'esecuzione. Puoi provarlo direttamente lanciando il seguente comando dal terminale:

```bash
go run examples/full_tutorial/main.go
```

Puoi ispezionare il codice sorgente completo dell'esempio nel file **[`examples/full_tutorial/main.go`](file:///home/tony/Projects/timescan/examples/full_tutorial/main.go)**.
