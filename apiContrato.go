package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/microsoft/go-mssqldb" // Driver SQL Server
)

// --- CAPA DE ENTIDAD (Entity) ---
type Contrato struct {
	ContratoNumero int       `json:"contrato_numero"`
	ClienteID      int       `json:"cliente_id"`
	FechaIngreso   time.Time `json:"fecha_ingreso"`
}

// --- CAPA DE REPOSITORIO (Repository) ---
type ContratoRepository interface {
	GetByID(ctx context.Context, id int) (*Contrato, error)
}

type sqlServerRepository struct {
	db *sql.DB
}

func NewContratoRepository(db *sql.DB) ContratoRepository {
	return &sqlServerRepository{db: db}
}

func (r *sqlServerRepository) GetByID(ctx context.Context, id int) (*Contrato, error) {
	// IMPORTANTE: Uso de QueryRowContext para manejo de timeouts y parámetros para evitar SQL Injection
	query := `
		SELECT Contrato_Numero, Cliente_Id, Fecha_Ingreso 
		FROM Contrato 
		WHERE Contrato_Id = @p1;`

	var c Contrato
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ContratoNumero, &c.ClienteID, &c.FechaIngreso)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No encontrado, no es un error técnico
		}
		return nil, err
	}
	return &c, nil
}

// --- CAPA DE TRANSPORTE (Handler) ---
type ContratoHandler struct {
	repo ContratoRepository
}

func NewContratoHandler(repo ContratoRepository) *ContratoHandler {
	return &ContratoHandler{repo: repo}
}

func (h *ContratoHandler) GetContrato(w http.ResponseWriter, r *http.Request) {
	// 1. Validación de Input
	idStr := r.URL.Query().Get("contrato_id")
	if idStr == "" {
		http.Error(w, "El parámetro contrato_id es requerido", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "El contrato_id debe ser un número entero válido", http.StatusBadRequest)
		return
	}

	// 2. Llamada al Repositorio con Contexto (Propaga cancelación si el cliente cierra conexión)
	contrato, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("Error en base de datos: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if contrato == nil {
		http.Error(w, "Contrato no encontrado", http.StatusNotFound)
		return
	}

	// 3. Respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contrato); err != nil {
		log.Printf("Error al codificar respuesta: %v", err)
	}
}

// --- CONFIGURACIÓN E INICIALIZACIÓN ---
func main() {
	// Configuración vía Variables de Entorno (12-Factor App)
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		"dbFortalezaCore",
	)

	// Conexión a Base de Datos
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatalf("Error creando pool de conexión: %v", err)
	}
	defer db.Close()

	// Validación de conexión (Healthcheck inicial)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}

	// Inyección de Dependencias
	repo := NewContratoRepository(db)
	handler := NewContratoHandler(repo)

	// Router (Usando Go 1.22+ mux estándar)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /contrato", handler.GetContrato)

	// Configuración del Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second, // Protección contra ataques Slowloris
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful Shutdown
	go func() {
		log.Printf("Servidor iniciado en puerto %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error en servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Apagando servidor...")

	ctxShut, cancelShut := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShut()
	if err := srv.Shutdown(ctxShut); err != nil {
		log.Fatal("Forzando apagado:", err)
	}
	log.Println("Servidor apagado correctamente.")
}
