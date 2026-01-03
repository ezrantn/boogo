// This AST encodes the Executable Boogie Subset (EBS v1).
// Constructs not representable here are intentionally excluded.

package boogie

// ========================
// Program Structure
// ========================

type Program struct {
	Procs []*Procedure
}

type Procedure struct {
	Name   string
	Params []Var
	Rets   []Var
	Locals []Var
	Body   []Stmt
}

// ========================
// Types (EBS v1)
// ========================

type Type interface {
	isType()
}

type IntType struct{}
type BoolType struct{}
type RefType struct{} // abstract heap reference

func (IntType) isType()  {}
func (BoolType) isType() {}
func (RefType) isType()  {}

// ========================
// Variables
// ========================

type Var struct {
	Name string
	Ty   Type
}

// ============================
// Statements (Structured Only)
// ============================

type Stmt interface {
	isStmt()
}

type LocalDecl struct {
	V Var
}

type Assume struct {
	Cond Expr
}

func (*Assume) isStmt() {}

func (*LocalDecl) isStmt() {}

type Assign struct {
	Lhs Expr
	Rhs Expr
}

func (*Assign) isStmt() {}

type If struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (*If) isStmt() {}

type While struct {
	Cond Expr
	Body []Stmt
}

func (*While) isStmt() {}

type Call struct {
	Name string
	Args []Expr
	Rets []Var // explicit return assignment
}

func (*Call) isStmt() {}

type Return struct {
	Values []Expr
}

func (*Return) isStmt() {}

// Verification-only (erasable)
type Assert struct {
	Cond Expr
}

func (*Assert) isStmt() {}

// ==================================
// Expressions (Pure & Deterministic)
// ==================================

type Expr interface {
	isExpr()
	Type() Type
}

// ---------- Variables ----------

type VarExpr struct {
	V Var
}

func (*VarExpr) isExpr() {}
func (e *VarExpr) Type() Type {
	return e.V.Ty
}

// ---------- Literals ----------

type IntLit struct {
	Value int
}

func (*IntLit) isExpr() {}
func (*IntLit) Type() Type {
	return IntType{}
}

type BoolLit struct {
	Value bool
}

func (*BoolLit) isExpr() {}
func (*BoolLit) Type() Type {
	return BoolType{}
}

// ---------- Binary Operations ----------

type BinOpKind int

const (
	Add BinOpKind = iota
	Sub
	Mul
	Eq
	Lt
	Lte
	Gt
	Gte
	And
	Or
)

type BinOp struct {
	Op    BinOpKind
	Left  Expr
	Right Expr
	Ty    Type
}

func (*BinOp) isExpr() {}
func (b *BinOp) Type() Type {
	return b.Ty
}

// ---------- Unary Operations ----------

type UnOpKind int

const (
	Not UnOpKind = iota
	Neg
)

type UnOp struct {
	Op UnOpKind
	X  Expr
	Ty Type
}

func (*UnOp) isExpr() {}
func (u *UnOp) Type() Type {
	return u.Ty
}

// ===========================
// Heap Operations (Restricted)
// ===========================

// HeapRead represents: Heap[o, f]
type HeapRead struct {
	Obj   Expr // must be RefType
	Field string
	Ty    Type
}

func (*HeapRead) isExpr() {}
func (*HeapRead) isStmt() {}
func (h *HeapRead) Type() Type {
	return h.Ty
}

// HeapWrite represents: Heap[o, f] := v
type HeapWrite struct {
	Obj   Expr // RefType
	Field string
	Value Expr
}

func (*HeapWrite) isStmt() {}
