package googlesheets

import (
	"context"
	"fmt"
	"time"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Exporter interface {
	GenerateInventoryReport(ctx context.Context, lots []domain.LotInternalResponse) (string, error)
	GeneratePnLReport(ctx context.Context, pnl *domain.PnLReport) (string, error)
}

type googleExporter struct {
	sheetsSrv     *sheets.Service
	spreadsheetID string
}

func NewGoogleExporter(ctx context.Context, credentialsFile string, spreadsheetID string) (Exporter, error) {
	opts := []option.ClientOption{
		option.WithCredentialsFile(credentialsFile),
		option.WithScopes(sheets.SpreadsheetsScope),
	}

	sheetsSrv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to init sheets service: %w", err)
	}

	return &googleExporter{
		sheetsSrv:     sheetsSrv,
		spreadsheetID: spreadsheetID,
	}, nil
}

// getOrCreateSheetId возвращает ID вкладки (SheetId нужен для форматирования)
func (e *googleExporter) getOrCreateSheetId(ctx context.Context, tabName string) (int64, error) {
	sp, err := e.sheetsSrv.Spreadsheets.Get(e.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return 0, err
	}

	for _, sheet := range sp.Sheets {
		if sheet.Properties.Title == tabName {
			return sheet.Properties.SheetId, nil
		}
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{Title: tabName},
			},
		}},
	}
	resp, err := e.sheetsSrv.Spreadsheets.BatchUpdate(e.spreadsheetID, req).Context(ctx).Do()
	if err != nil {
		return 0, err
	}
	return resp.Replies[0].AddSheet.Properties.SheetId, nil
}

// applyFormatting наводит красоту: рамки, центрирование, жирный шрифт и автоширина
func (e *googleExporter) applyFormatting(ctx context.Context, sheetId int64, numRows int64, numCols int64, boldRows []int64, boldCols []int64) error {
	var requests []*sheets.Request

	// 1. Выравнивание текста по центру + ДОБАВЛЯЕМ ВОЗДУХ (Padding)
	requests = append(requests, &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetId,
				StartRowIndex:    0,
				EndRowIndex:      numRows,
				StartColumnIndex: 0,
				EndColumnIndex:   numCols,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					HorizontalAlignment: "CENTER",
					VerticalAlignment:   "MIDDLE",
					Padding: &sheets.Padding{
						Left:   12, // Добавляем 12 пикселей слева
						Right:  12, // Добавляем 12 пикселей справа
						Top:    6,
						Bottom: 6,
					},
				},
			},
			// Обязательно указываем padding в Fields, иначе Гугл его проигнорирует
			Fields: "userEnteredFormat(horizontalAlignment,verticalAlignment,padding)",
		},
	})

	// 2. Рисуем сетку (рамки)
	border := &sheets.Border{Style: "SOLID", Color: &sheets.Color{}}
	requests = append(requests, &sheets.Request{
		UpdateBorders: &sheets.UpdateBordersRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetId,
				StartRowIndex:    0,
				EndRowIndex:      numRows,
				StartColumnIndex: 0,
				EndColumnIndex:   numCols,
			},
			Top:             border,
			Bottom:          border,
			Left:            border,
			Right:           border,
			InnerHorizontal: border,
			InnerVertical:   border,
		},
	})

	// 3. Делаем указанные строки жирными
	for _, rIdx := range boldRows {
		requests = append(requests, &sheets.Request{
			RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetId,
					StartRowIndex:    rIdx,
					EndRowIndex:      rIdx + 1,
					StartColumnIndex: 0,
					EndColumnIndex:   numCols,
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{TextFormat: &sheets.TextFormat{Bold: true}},
				},
				Fields: "userEnteredFormat.textFormat.bold",
			},
		})
	}

	// 4. Делаем указанные колонки жирными
	for _, cIdx := range boldCols {
		requests = append(requests, &sheets.Request{
			RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetId,
					StartRowIndex:    0,
					EndRowIndex:      numRows,
					StartColumnIndex: cIdx,
					EndColumnIndex:   cIdx + 1,
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{TextFormat: &sheets.TextFormat{Bold: true}},
				},
				Fields: "userEnteredFormat.textFormat.bold",
			},
		})
	}

	// 5. Автоматическая ширина колонок под длину текста (с учетом новых отступов)
	requests = append(requests, &sheets.Request{
		AutoResizeDimensions: &sheets.AutoResizeDimensionsRequest{
			Dimensions: &sheets.DimensionRange{
				SheetId:    sheetId,
				Dimension:  "COLUMNS",
				StartIndex: 0,
				EndIndex:   numCols, // ИСПРАВЛЕНО: убрали int32(), теперь типы совпадают
			},
		},
	})

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{Requests: requests}
	_, err := e.sheetsSrv.Spreadsheets.BatchUpdate(e.spreadsheetID, batchReq).Context(ctx).Do()
	return err
}

func (e *googleExporter) GenerateInventoryReport(ctx context.Context, lots []domain.LotInternalResponse) (string, error) {
	tabName := fmt.Sprintf("Остатки %s", time.Now().Format("02.01 15:04:05"))
	sheetId, err := e.getOrCreateSheetId(ctx, tabName)
	if err != nil {
		return "", err
	}

	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{"Сгенерировано:", time.Now().Format("02.01.2006 15:04:05"), "", "", "", "", "", "", ""})
	vr.Values = append(vr.Values, []interface{}{"ID Лота", "Тип", "Состояние", "Бренд", "Модель", "Остаток (шт)", "Закупка ($)", "Продажа ($)", "Статус"})

	for _, lot := range lots {
		vr.Values = append(vr.Values, []interface{}{
			lot.ID.String(), lot.Type, lot.Condition, lot.Brand, lot.Model,
			lot.CurrentQuantity, lot.PurchasePrice, lot.SellPrice, lot.Status,
		})
	}

	// Запись данных
	writeRange := fmt.Sprintf("'%s'!A1", tabName)
	_, err = e.sheetsSrv.Spreadsheets.Values.Update(e.spreadsheetID, writeRange, &vr).
		ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to write data: %w", err)
	}

	// Применяем форматирование
	// numRows = len(vr.Values), numCols = 9
	// boldRows = [1] (Заголовки), boldCols = [0] (ID Лота)
	e.applyFormatting(ctx, sheetId, int64(len(vr.Values)), 9, []int64{1}, []int64{0})

	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", e.spreadsheetID), nil
}

func (e *googleExporter) GeneratePnLReport(ctx context.Context, pnl *domain.PnLReport) (string, error) {
	tabName := fmt.Sprintf("P&L %s", time.Now().Format("02.01 15:04:05"))
	sheetId, err := e.getOrCreateSheetId(ctx, tabName)
	if err != nil {
		return "", err
	}

	var vr sheets.ValueRange

	// Строка 0: Дата
	vr.Values = append(vr.Values, []interface{}{"Сгенерировано:", time.Now().Format("02.01.2006 15:04:05"), "", "", ""})
	// Строка 1: Заголовки
	vr.Values = append(vr.Values, []interface{}{"Локация", "Продано (шт)", "Выручка (Revenue) $", "Себестоимость (COGS) $", "Чистая Прибыль (Profit) $"})
	// Строка 2: Итого (Без эмодзи)
	vr.Values = append(vr.Values, []interface{}{
		"ИТОГО ПО КОМПАНИИ",
		pnl.TotalItemsSold,
		pnl.TotalRevenue,
		pnl.TotalCOGS,
		pnl.TotalProfit,
	})

	// Строки 3+: Склады (убрал префикс "Склад: ", так как колонка называется "Локация")
	for _, w := range pnl.ByWarehouse {
		vr.Values = append(vr.Values, []interface{}{
			w.WarehouseName,
			w.ItemsSold,
			w.Revenue,
			w.COGS,
			w.Profit,
		})
	}

	// Запись данных
	writeRange := fmt.Sprintf("'%s'!A1", tabName)
	_, err = e.sheetsSrv.Spreadsheets.Values.Update(e.spreadsheetID, writeRange, &vr).
		ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to write data: %w", err)
	}

	// Применяем форматирование
	// numRows = len(vr.Values), numCols = 5
	// boldRows = [1, 2] (Заголовки и строка "ИТОГО"), boldCols = [0] (Названия складов)
	e.applyFormatting(ctx, sheetId, int64(len(vr.Values)), 5, []int64{1, 2}, []int64{0})

	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", e.spreadsheetID), nil
}
