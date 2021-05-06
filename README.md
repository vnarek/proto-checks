# Výstižný nadpis
## Zadání

## Motivace

## Implementace
TODO: nějaký načrt rozčlenění na 3 části - sestavení CFG, pointer analýza a nilness analýza

### Control flow graph
Jako první jsme museli navrhnout, jak bude vypadat control flow graph (CFG). Obě dvě analýzy vyžadují na vstupu program, ve kterém jsou normalizované všechny operaci s ukazeteli do následujích tvarů:
```c++
X = alloc P
X1 = &X2
X1 = X2
X1 = *X2
X = null (nil v případě jazyka Go)
```
Náš CFG se tedy narozdíl od běžných CFG skládá pouze z těchto šesti uzlů (ostatní operace nejsou pro naše potřeby důležité, proto je z grafu vynecháme).
Dále pro účely null pointer analýzy jsme od grafu vyžadovali, aby každý uzel obsahoval jak hrany na svoje následovníky,
tak i na své předky. Jelikož tyto požadavky byly velmi konkrétní, bylo nám téměr jasné, že CFG si budeme muset napsat z velké části sami.

K sestavení CFG jsme využili [go knihovnu cfg](https://pkg.go.dev/golang.org/x/tools/go/cfg), která pro každou funkci
sestaví graf basic bloků (ty už neobsahují AST nody, které mění plynutí programu - tzn. If, Switch, atd.).
Bohužel tento graf je jednosměrný, pro každý uzel jsou definovaný jen následovníci. Předky jsme tedy museli doplnit.
Dále jsme z bloku odstranili veškeré AST nody, které neobsahují operace s ukazateli. To znamená, že některé basic bloky
z grafu vypadly úplně. V tom případě bylo potřeba správně napojit hrany grafu, aby se neporušila struktura programu. Např. tento Go program:
```go
sum := new(int)
for i := 0; i < 2; i++ {
    sum = x
}
res := *sum
```
Knihovna cfg rozloží do těchto pěti basic bloků:
```
0:  (next: 3)
    sum := new(int)
    i := 0
    
1:  (next: 4)
    sum = **x
    
2:  (next: )
    res := *sum
    
3:  (next 1, 2)
    i < 2
    
4:  (next: 3)
    i++
```
Nyní spustíme náš algoritmus, který začne normalizovat od prvního basic bloku jednotlivé výrazy a zároveň se začne zbavovat výrazů, které pro nás
nejsou důležité. Z příkladu si můžeme všimnout, že bloky 3 a 4 neobsahují žádné operace s ukazeteli. Těchto bloků
se zbavíme úplně a správně napojíme hrany. Zároveň doplníme kromě následovníků i hrany na předchůdce:
```
0:  (next: 1, 2) (prev: )
    sum := new(int)
    
1:  (next: 1, 2) (prev: 0, 1)
    sum = **x
    
2:  (next: ) (prev: 0, 1)
    res := *sum
```
K normalizaci výrazu dojde v případě uzlu 1 (protože nemá jeden z 6 tvarů, který obě analýzy vyžadují), ze kterého se
stanou dva uzly, který již požadovaný tvar budou mít:
```
0:  (next: 1_1, 2) (prev: )
    sum := new(int)
    
1_1:(next: 1_2) (prev: 0, 1_2)
    _t1 = *x
    
1_2:(next: 1_1, 2) (prev: 1_1)
    sum = *_t1
    
2:  (next: ) (prev: 0, 1_2)
    res := *sum
```
Výsledkem našeho algoritmu je tedy následující CFG:
```
     ↓
[sum = alloc]→→
     ↓        ↓
[_t1 = *x]    ↓
    ↓ ↑       ↓
[sum = *_t1]  ↓
     ↓        ↓
[res = *sum]←←↓
     ↓
```
Tento graf používá null pointer analýza. Pointer analýza používá modifikovanou verzi tohoto grafu, která se skládá
z po sobě jdoucích uzlech (nejedná se tedy úplně o graf, jako spíš sekvenci uzlů). Je to kvůli tomu, že k analýze se
používá Andesenův algoritmus, který je flow-insensitive a nezáleží mu na pořadí jednotlivých uzlů.

Algoritmus se nachází v souboru `cfg/cfg.go`. Dále se nachází v `cfg/testdata` sada testů, kde referenční výstupy
jednotlivých casů jsou uloženy v souborech s příponou *.golden*. Tyto testy většinou kontrolují správnou normalizaci
operací s ukazateli, zároveň se i snaží pokrýt všechny způsoby, jak jdou jednotlivé operace definovat. Pro zajímavost,
například uzel `[x = null]` jde v jazyku Go zapsat těmito ekvivalentními způsoby:
```go
x = nil
x := nil
var x *int = nil
var x *int
x = (((nil)))
```
Celá sada testů se spouští ze složky `cfg` příkazem `go test -v ./...`.

### Pointer analýza

### Null pointer analýza

## Závěr
