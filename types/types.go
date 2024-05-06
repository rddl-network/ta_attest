package types

type PostCreateAccountRequest struct {
	MachineID    string `binding:"required" json:"machine-id"`
	PlmntAddress string `binding:"required" json:"plmnt-address"`
	Signature    string `binding:"required" json:"signature"`
}
