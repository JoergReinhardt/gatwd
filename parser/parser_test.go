package parser

//func TestRuneText(t *testing.T) {
//	var text = newRuneText("this is not a test")
//	fmt.Printf("allocated rune slice: %s\n", text)
//
//	var head, _ = text.Head()
//	fmt.Printf("head rune slice: %s\n", string(head))
//	fmt.Printf("rune slice after inspecting head: %s\n", text)
//
//	var ok bool
//	head, text, ok = text.Consume()
//	fmt.Printf("consume rune slice, head: %s, ok: %t, remains: %s\n",
//		string(head), ok, text)
//
//	for ok {
//		fmt.Printf("loop consume rune slice,"+
//			" head: %s, ok: %t, remains: %s\n",
//			string(head), ok, text)
//		head, text, ok = text.Consume()
//	}
//
//	var text1 = newRuneText("also not a test")
//	var took = []rune{}
//
//	took, text1, ok = text1.Take(4)
//	fmt.Printf("took four from rune slice,"+
//		" head: %s, ok: %t, remains: %s\n",
//		string(took), ok, text1)
//
//	var peek rune
//	peek, ok = text1.Peek(3)
//	fmt.Printf("peek index 3 of rune slice (should be 't'),"+
//		" head: %s, ok: %t, remains: %s\n",
//		string(peek), ok, text1)
//
//	peek, ok = text1.Peek(11)
//	fmt.Printf("peek index 11 of rune slice (should fail),"+
//		" head: %s, ok: %t, remains: %s\n",
//		string(peek), ok, text1)
//
//	took, text1, ok = text1.Take(11)
//	fmt.Printf("try take til 11 of rune slice (should fail),"+
//		" took: %s, ok: %t, remains: %s\n",
//		string(took), ok, text1)
//
//	took, text1, ok = text1.Take(10)
//	fmt.Printf("try take til 10 of rune slice (should succeed but be false "+
//		"remain should be empty), took: %s, ok: %t, remains: %s\n",
//		string(took), ok, text1)
//}
