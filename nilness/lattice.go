package nilness

import "github.com/vnarek/proto-checks/cfg"

type NilVal int

const (
	NN NilVal = iota
	PN
)

//NilnessLattice represents flat lattice of nilness analysis
type NilnessLattice struct {

}

func (NilnessLattice) Top() NilVal {
	return PN
}

func (NilnessLattice) Bot() NilVal {
	return NN
}

func (NilnessLattice) Lub(x NilVal, y NilVal) NilVal {
	if x == NN && y == NN {
		return NN
	}
	return PN
}


//DeclMapLattice represents MapLattice[string, NilnessLattice]
type DeclMapLattice struct {
	lattice NilnessLattice
}

func NewDeclMapLattice() DeclMapLattice {
	return DeclMapLattice{lattice: NilnessLattice{}}
}

type DeclMap struct {
	DefaultVal NilVal
	Data map[string]NilVal
}

func (m DeclMap) Get(k string) NilVal {
	v, ok := m.Data[k]
	if ok {
		return v
	}
	return m.DefaultVal
}

func (m DeclMap) Set(k string, v NilVal) {
	m.Data[k] = v
}

func (m DeclMap) Clone() DeclMap {
	newMap := NewDeclMap(m.DefaultVal)
	for k, v := range m.Data {
		newMap.Set(k, v)
	}
	return newMap
}

func NewDeclMap(defaultVal NilVal) DeclMap {
	return DeclMap{
		DefaultVal: defaultVal,
		Data:       make(map[string]NilVal),
	}
}

func (DeclMapLattice) Top() DeclMap {
	return NewDeclMap(PN)
}

func (DeclMapLattice) Bot() DeclMap {
	return NewDeclMap(NN)
}

func (d DeclMapLattice) Lub(x DeclMap, y DeclMap) DeclMap {
	res := y.Clone()
	for k := range x.Data {
		res.Set(k, d.lattice.Lub(x.Get(k), y.Get(k)))
	}
	return res
}

func (m DeclMap) Eq(other DeclMap) bool {
	for k, v := range m.Data {
		otherV := other.Get(k)
		if v != otherV {
			return false
		}
	}
	for k, v := range other.Data {
		otherV := m.Get(k)
		if v != otherV {
			return false
		}
	}
	return true
}








//CfgMapLattice represents MapLattice[cfg.Node, DeclMapLattice]
type CfgMapLattice struct {
	lattice DeclMapLattice
}

func NewCfgMapLattice(lattice DeclMapLattice) CfgMapLattice {
	return CfgMapLattice{lattice: lattice}
}

type CfgMap struct {
	DefaultVal DeclMap
	Data map[cfg.Node]DeclMap
}

func (m CfgMap) Get(k cfg.Node) DeclMap {
	v, ok := m.Data[k]
	if ok {
		return v
	}
	return m.DefaultVal
}

func (m CfgMap) Set(k cfg.Node, v DeclMap) {
	m.Data[k] = v
}

func (m CfgMap) Clone() CfgMap {
	newMap := NewCfgMap(m.DefaultVal)
	for k, v := range m.Data {
		newMap.Set(k, v)
	}
	return newMap
}

func (m CfgMap) Eq(other CfgMap) bool {
	for k, v := range m.Data {
		otherV := other.Get(k)
		if !v.Eq(otherV) {
			return false
		}
	}
	for k, v := range other.Data {
		otherV := m.Get(k)
		if !v.Eq(otherV) {
			return false
		}
	}
	return true
}

func NewCfgMap(defaultVal DeclMap) CfgMap {
	return CfgMap{
		DefaultVal: defaultVal,
		Data:       make(map[cfg.Node]DeclMap),
	}
}

func (c CfgMapLattice) Top() CfgMap {
	return NewCfgMap(c.lattice.Top())
}

func (c CfgMapLattice) Bot() CfgMap {
	return NewCfgMap(c.lattice.Bot())
}

func (c CfgMapLattice) Lub(x CfgMap, y CfgMap) CfgMap {
	res := y.Clone()
	for k, _ := range x.Data {
		res.Set(k, c.lattice.Lub(x.Get(k), y.Get(k)))
	}
	return res
}


