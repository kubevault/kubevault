package metrics

import "gopkg.in/square/go-jose.v2/jwt"

// TODO: is it stable
// TODO: change it
type License struct {
	Issuer    string           `json:"issuer,omitempty"`     // byte.builders
	Subject   string           `json:"subject,omitempty"`    // user_id
	Audience  jwt.Audience     `json:"audience,omitempty"`   // cluster_id ?
	Expiry    *jwt.NumericDate `json:"expiry,omitempty"`     // if set, use this
	NotBefore *jwt.NumericDate `json:"not_before,omitempty"` // start of subscription start
	IssuedAt  *jwt.NumericDate `json:"issued_at,omitempty"`  // timestamp of issue
	ID        string           `json:"id,omitempty"`         // license ID from firestore
	Owner     string           `json:"owner,omitempty"`      // product owner / produce: appscode
	Status    string           `json:"status"`
}
