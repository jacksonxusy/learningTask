package main

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//编写一个事务，实现从账户 A 向账户 B 转账 100 元的操作。在事务中，需要先检查账户 A 的余额是否足够，
//如果足够则从账户 A 扣除 100 元，
//向账户 B 增加 100 元，并在 transactions 表中记录该笔转账信息。如果余额不足，则回滚事务。

type Accounts struct {
	ID      int
	Balance int
}

type Transaction struct {
	ID            int
	FromAccountID int
	ToAccountID   int
	Amount        int
}

func main() {
	db, err := gorm.Open(mysql.Open("root:root1234@tcp(127.0.0.1:3306)/gorm?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	accountsA := Accounts{ID: 1, Balance: 1000}
	accountsB := Accounts{ID: 2, Balance: 1000}

	db.AutoMigrate(&Accounts{}, &Transaction{})
	db.Create(&accountsA)
	db.Create(&accountsB)
	fromAccountID := accountsA.ID
	toAccountID := accountsB.ID
	amount := 100
	db.Transaction(func(tx *gorm.DB) error {
		// 检查账户 A 的余额是否足够
		tx.First(&accountsA, fromAccountID)
		if accountsA.Balance < 100 {
			return errors.New("insufficient balance")
		}
		// 从账户 A 扣除 100 元
		tx.Model(&Accounts{}).Where("id = ?", fromAccountID).Update("balance", accountsA.Balance-amount)

		// 向账户 B 增加 100 元
		tx.Model(&Accounts{}).Where("id = ?", toAccountID).Update("balance", gorm.Expr("balance + ?", amount))

		// 记录转账信息
		tx.Create(&Transaction{
			FromAccountID: fromAccountID,
			ToAccountID:   toAccountID,
			Amount:        amount,
		})

		return nil
	})
}
