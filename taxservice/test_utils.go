package taxservice

import (
	"context"
	"github.com/rezkam/TaxMan/model"
)

type mockStore struct {
	addOrUpdateTaxRecordFunc func(ctx context.Context, record model.TaxRecord) error
	getTaxRecordsFunc        func(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error)
}

func (m *mockStore) AddOrUpdateTaxRecord(ctx context.Context, record model.TaxRecord) error {
	if m.addOrUpdateTaxRecordFunc != nil {
		return m.addOrUpdateTaxRecordFunc(ctx, record)
	}
	return nil
}

func (m *mockStore) GetTaxRecords(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error) {
	if m.getTaxRecordsFunc != nil {
		return m.getTaxRecordsFunc(ctx, query)
	}
	return nil, nil
}
