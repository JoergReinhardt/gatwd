/*
 CATEGORY OBJECT INTERFACE AND UNIT TYPES

Clearup Category

Category is cluttered by dispatch of lable/id/flag accessors.
Category will only be accessed by index via 'Obj.Id() -> int'.
  ⇒ category constructor needs to derive Kind from int Id:
    ⇒ kind needs to be kind[0]
      ⇒ kind needs to return all []kind
        ← in order to access sibling kinds
      ⇒ kind needs to return its kind: kind[0].Kind()
        ← in order to walk tree for generation of uuid

  '''
  Category  ∷ C  →  Obj…

    id   C  = id    C →  int		   |  pos in parent category
    card C  = card  C →  int		   |  length
    name C  = name  C →  string		   |  name of category (flag lable, or symbol)
					   |
    cons    = C  →  C₀			   |  cons empty root category
    cons O… = O… →  (O.Id()… → Cₙ) → Cₙ₊ₘ  |  cons 'cat from cat & objects
					   |
    kind    = C  →  Cₙ…			   |  kind ∅ returns all kinds	    → C…
    kind Cₙ = Cₙ →  (Cₙ → C₀) → Cₚₐᵣₑₙₜ	   |  kind C returns kind <of> O(C) → Cₚ
    kind Oₙ = Oₙ →  (id Oₙ → Cₙ) → C₍ₕᵢₗ₎  |  kind O returns kind <of> O₍   → Cₒ
  '''

 every thing is an object and needs to implement the object interface 'Obj'.
 that includes internal parts of the type system, like type markers and
 names.  the interface demands a 'Type() int' method to return a unique
 numeric identification.

 the 'Ident() Obj' method needs to be implemented to return the native
 instance of whatever type implements the interface, aka it-'self'.

   - runtime defined types are accessed by slice index.

   - category types need quick set membership identification

   - some kinds are named, others need anonymity.

 hence three sorts of identity markers exist:

   - numeric unique id shared by every kind of type.

   - binary bit flag for sets of categorys, with quick membership operation

   - string representation of instance data, or name of its type

      CATEGORY ∷ C  →  [O]			   <|>  set of objects
      cons	    = ∅  →  C₀			   <|>  cons all sub categorys
      cons	C   = C  →  [C]			   <|>  cons parent category
      cons O…  = O… →  (O.Id()… → Cₙ) → Cₙ₊ₘ  <|>  cons new cat' from cat & objects
					  <|>
      kind    = C  →  Cₙ…			   <|>  kind ∅ returns all kinds      → C…
      kind Cₙ = Cₙ →  (Cₙ → C₀) → Cₚₐᵣₑₙₜ	   <|>  kind C returns kind <of> O(C) → Cₚ
      kind Oₙ = Oₙ →  (id Oₙ → Cₙ) → C₍ₕᵢₗ₎   <|>  kind O returns kind <of> O₍   → Cₒ
      ⇒ (kind C →  Cₚₐᵣₑₙₜ| kind O →  C₍ₕᵢₗ₎)
      ⇒ O… <cons> C  →  C'

	//Sum Obj // ∑ [Objₜ] = collection type of particular objects type
	//Pro Obj // ∏ Obj₁|Obj₂|…|Objₙ = enum type composed of n subtype flags

//	Lnk interface {
//		Head() Obj
//		Tail() Obj
//		Ident
//	}
//
//	Cnst func() Ident          // constant object
//	Pair func() (Ident, Ident) // pair of objects
//	Link func() (Ident, Pair)  // linked objects (list, tree, …)
//	Vect []Ident               // sum of objects
//
//	UnaOp func(Ident) Ident      // unary operation
//	BinOp func(a, b Ident) Ident // binary operation
//	GenOp func(...Ident) Ident   // generic operation
*/
package gatw
