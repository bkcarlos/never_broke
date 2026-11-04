package models

import (
	"time"

	"gorm.io/gorm"
)

// Account GORM 账户模型
type Account struct {
	ID          string   `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string   `gorm:"type:varchar(36);not null;index:idx_user_accounts" json:"userId"`
	Type        string   `gorm:"type:varchar(20);not null" json:"type"`
	Institution *string  `gorm:"type:varchar(50)" json:"institution,omitempty"`
	CardLast4   *string  `gorm:"type:varchar(4)" json:"cardLast4,omitempty"`
	Currency    string   `gorm:"type:varchar(10);not null" json:"currency"`
	Balance     float64  `gorm:"type:decimal(15,2);not null;default:0" json:"balance"`
	CreditLimit *float64 `gorm:"type:decimal(15,2)" json:"creditLimit,omitempty"`
	TempLimit   *float64 `gorm:"type:decimal(15,2)" json:"tempLimit,omitempty"`
	Archived    bool     `gorm:"not null;default:false" json:"archived"`
}

// TableName 指定表名
func (Account) TableName() string {
	return "accounts"
}

// Transaction GORM 交易模型
type Transaction struct {
	ID                string  `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID            string  `gorm:"type:varchar(36);not null;index:idx_user_transactions" json:"userId"`
	Type              string  `gorm:"type:varchar(20);not null" json:"type"`
	Amount            float64 `gorm:"type:decimal(15,2);not null" json:"amount"`
	CategoryKey       *string `gorm:"type:varchar(50)" json:"categoryKey,omitempty"`
	Remark            *string `gorm:"type:text" json:"remark,omitempty"`
	OccurredAt        string  `gorm:"type:varchar(30);not null;index:idx_occurred_at" json:"occurredAt"`
	AccountID         *string `gorm:"type:varchar(36);index:idx_account" json:"accountId,omitempty"`
	FromAccountID     *string `gorm:"type:varchar(36)" json:"fromAccountId,omitempty"`
	ToAccountID       *string `gorm:"type:varchar(36)" json:"toAccountId,omitempty"`
	InstallmentPlanID *string `gorm:"type:varchar(36);index:idx_installment" json:"installmentPlanId,omitempty"`
	InstallmentPeriod *int    `gorm:"type:int" json:"installmentPeriod,omitempty"`
	InstallmentTotal  *int    `gorm:"type:int" json:"installmentTotal,omitempty"`
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transactions"
}

// InstallmentPlan GORM 分期计划模型
type InstallmentPlan struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID        string    `gorm:"type:varchar(36);not null;index:idx_user_installments" json:"userId"`
	AccountID     string    `gorm:"type:varchar(36);not null;index:idx_account_installments" json:"accountId"`
	Principal     float64   `gorm:"type:decimal(15,2);not null" json:"principal"`
	TotalInterest float64   `gorm:"type:decimal(15,2);not null" json:"totalInterest"`
	TotalAmount   float64   `gorm:"type:decimal(15,2);not null" json:"totalAmount"`
	TermMonths    int       `gorm:"type:int;not null" json:"termMonths"`
	MonthlyRate   float64   `gorm:"type:decimal(10,6);not null" json:"monthlyRate"`
	RepayMethod   string    `gorm:"type:varchar(30);not null" json:"repayMethod"`
	Currency      string    `gorm:"type:varchar(10);not null" json:"currency"`
	FirstDueDate  time.Time `gorm:"type:date;not null" json:"firstDueDate"`
	Status        string    `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt     time.Time `gorm:"type:timestamptz;not null" json:"createdAt"`
}

// TableName 指定表名
func (InstallmentPlan) TableName() string {
	return "installment_plans"
}

// RecurringIncomePlan GORM 周期收入计划模型
type RecurringIncomePlan struct {
	ID                    string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID                string     `gorm:"type:varchar(36);not null;index:idx_user_recurring" json:"userId"`
	Type                  string     `gorm:"type:varchar(20);not null" json:"type"`
	Label                 string     `gorm:"type:varchar(100);not null" json:"label"`
	AccountID             string     `gorm:"type:varchar(36);not null" json:"accountId"`
	Amount                float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	Currency              string     `gorm:"type:varchar(10);not null" json:"currency"`
	Frequency             string     `gorm:"type:varchar(20);not null" json:"frequency"`
	DayOfMonth            *int       `gorm:"type:int" json:"dayOfMonth,omitempty"`
	Weekday               *int       `gorm:"type:int" json:"weekday,omitempty"`
	NextDueDate           time.Time  `gorm:"type:date;not null" json:"nextDueDate"`
	LastFiredDate         *time.Time `gorm:"type:date" json:"lastFiredDate,omitempty"`
	CityCode              *string    `gorm:"type:varchar(50)" json:"cityCode,omitempty"`
	Active                bool       `gorm:"not null;default:true" json:"active"`
	CreatedAt             time.Time  `gorm:"type:timestamptz;not null" json:"createdAt"`
	SalaryContributions   *string    `gorm:"type:text" json:"salaryContributions,omitempty"`
	SalaryPreTaxMonthly   *float64   `gorm:"type:decimal(15,2)" json:"salaryPreTaxMonthly,omitempty"`
	SalarySocialRate      *float64   `gorm:"type:decimal(5,4)" json:"salarySocialRate,omitempty"`
	SalaryExtraDeductions *float64   `gorm:"type:decimal(15,2)" json:"salaryExtraDeductions,omitempty"`
	SalaryPayDay          *int       `gorm:"type:int" json:"salaryPayDay,omitempty"`
	SalaryStartDate       *time.Time `gorm:"type:date" json:"salaryStartDate,omitempty"`
}

// TableName 指定表名
func (RecurringIncomePlan) TableName() string {
	return "recurring_income_plans"
}

// Budget GORM 预算模型
type Budget struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string    `gorm:"type:varchar(36);not null;uniqueIndex:idx_user_month" json:"userId"`
	Month       string    `gorm:"type:varchar(7);not null;uniqueIndex:idx_user_month" json:"month"` // YYYY-MM
	TotalBudget float64   `gorm:"type:decimal(15,2);not null;default:0" json:"totalBudget"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null" json:"updatedAt"`
}

// TableName 指定表名
func (Budget) TableName() string {
	return "budgets"
}

// IncomePlan GORM 收入计划模型
type IncomePlan struct {
	ID           string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID       string     `gorm:"type:varchar(36);not null;index:idx_income_plans_user" json:"userId"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	Amount       float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	Frequency    string     `gorm:"type:varchar(20);not null" json:"frequency"`
	AccountID    string     `gorm:"type:varchar(36);not null" json:"accountId"`
	FirstPayDate time.Time  `gorm:"type:date;not null" json:"firstPayDate"`
	EndDate      *time.Time `gorm:"type:date" json:"endDate,omitempty"`
	Active       bool       `gorm:"not null;default:true" json:"active"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;not null" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;not null" json:"updatedAt"`
}

// TableName 指定表名
func (IncomePlan) TableName() string {
	return "income_plans"
}

// BeforeCreate GORM hook - 在创建前设置时间
func (a *Account) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeCreate GORM hook - 在创建前设置时间
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeCreate GORM hook - 在创建前设置时间
func (i *InstallmentPlan) BeforeCreate(tx *gorm.DB) error {
	if i.CreatedAt.IsZero() {
		i.CreatedAt = time.Now()
	}
	return nil
}

// BeforeCreate GORM hook - 在创建前设置时间
func (r *RecurringIncomePlan) BeforeCreate(tx *gorm.DB) error {
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	return nil
}

// BeforeCreate GORM hook - 在创建前设置时间
func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate GORM hook - 在更新前设置时间
func (b *Budget) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate GORM hook - 在创建前设置时间
func (p *IncomePlan) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate GORM hook - 在更新前设置时间
func (p *IncomePlan) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}
