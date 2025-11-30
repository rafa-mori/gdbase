// Package clients fornece funcionalidades para gerenciar clientes.
package clients

import (
	"context"
	"fmt"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"
	"gorm.io/gorm"
)

// IClientRepo define o contrato do repositório para clientes.
type IClientRepo interface {
	// Cria um novo cliente e retorna o objeto criado.
	Create(client *ClientDetailed) (*ClientDetailed, error)
	// Busca um cliente usando uma condição genérica (Ex.: "id = ?", id).
	FindOne(query interface{}, args ...interface{}) (*ClientDetailed, error)
	// Busca todos os clientes que satisfaçam determinada condição.
	FindAll(query interface{}, args ...interface{}) ([]*ClientDetailed, error)
	// Atualiza os dados de um cliente.
	Update(client *ClientDetailed) (*ClientDetailed, error)
	// Exclui um cliente com base no ID.
	Delete(id string) error
	// Fecha a conexão com o banco de dados.
	Close() error
	// Lista os clientes em um formato de tabela simples ou outro formato que desejar.
	List(query interface{}, args ...interface{}) (interface{}, error)
}

// ClientRepo é a implementação de IClientRepo usando GORM.
type ClientRepo struct {
	db *gorm.DB
}

// NewClientRepo cria uma nova instância de ClientRepo.
func NewClientRepo(ctx context.Context, dbService *svc.DBServiceImpl) IClientRepo {
	if dbService == nil {
		gl.Log("error", "ClientRepo: dbService is nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("ClientRepo: failed to get DB from dbService: %v", err))
		return nil
	}
	return &ClientRepo{db: db}
}

func (cr *ClientRepo) Create(client *ClientDetailed) (*ClientDetailed, error) {
	if client == nil {
		return nil, fmt.Errorf("ClientRepo: client is nil")
	}
	err := cr.db.Create(client).Error
	if err != nil {
		return nil, fmt.Errorf("ClientRepo: failed to create client: %w", err)
	}
	return client, nil
}

func (cr *ClientRepo) FindOne(query interface{}, args ...interface{}) (*ClientDetailed, error) {
	var client ClientDetailed
	err := cr.db.Where(query, args...).First(&client).Error
	if err != nil {
		return nil, fmt.Errorf("ClientRepo: failed to find client: %w", err)
	}
	return &client, nil
}

func (cr *ClientRepo) FindAll(query interface{}, args ...interface{}) ([]*ClientDetailed, error) {
	var clients []*ClientDetailed
	err := cr.db.Where(query, args...).Find(&clients).Error
	if err != nil {
		return nil, fmt.Errorf("ClientRepo: failed to find clients: %w", err)
	}
	return clients, nil
}

func (cr *ClientRepo) Update(client *ClientDetailed) (*ClientDetailed, error) {
	if client == nil {
		return nil, fmt.Errorf("ClientRepo: client is nil")
	}
	err := cr.db.Save(client).Error
	if err != nil {
		return nil, fmt.Errorf("ClientRepo: failed to update client: %w", err)
	}
	return client, nil
}

func (cr *ClientRepo) Delete(id string) error {
	err := cr.db.Delete(&ClientDetailed{}, id).Error
	if err != nil {
		return fmt.Errorf("ClientRepo: failed to delete client: %w", err)
	}
	return nil
}

func (cr *ClientRepo) Close() error {
	sqlDB, err := cr.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (cr *ClientRepo) List(query interface{}, args ...interface{}) (interface{}, error) {
	var clients []ClientDetailed
	err := cr.db.Where(query, args...).Find(&clients).Error
	if err != nil {
		return nil, fmt.Errorf("ClientRepo: failed to list clients: %w", err)
	}
	// Aqui, por exemplo, podemos construir uma estrutura de tabela simples
	tableRows := [][]string{}
	for i, client := range clients {
		row := []string{
			fmt.Sprintf("%d", i+1),
			client.ID,
			// Supondo que TradingName seja opcional; usamos uma função inline para tratar o nil
			func() string {
				if client.TradingName != nil {
					return *client.TradingName
				}
				return ""
			}(),
			string(client.Status), // status é do tipo ClientStatus (string)
		}
		tableRows = append(tableRows, row)
	}
	return tableRows, nil
}
