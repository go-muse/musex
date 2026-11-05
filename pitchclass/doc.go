// Package pitchclass provides pitch classes: the positions within an octave
// of a system with N positions, counted from 0, together with arithmetic on
// them modulo N, multipliers, operators x ↦ m·x + n, groups of operators,
// and their text forms.
//
// A [Class] is a class k of N positions, written "k mod N". An [Operator] is
// a transformation x ↦ m·x + n (mod N) of the classes of one N. A [Style] is
// a style of the literary notation of class numbers.
//
// The package defines the following operations:
//
//   - [New] and [Reduce] create a class strictly or by wrapping any integer
//     around the circle; [All] lists the classes of N; [Class.Number],
//     [Class.Modulus], and [Class.IsValid] read a class; [Class.Equal] and
//     [Class.Compare] compare classes; [Class.Cycle] walks the circle upward
//     from a class.
//   - [Class.Transpose] transposes a class; [Class.IntervalTo],
//     [Class.IntervalClass], and [Class.IndexSum] measure a pair of classes;
//     [ReduceInterval], [ComplementInterval], and [IntervalClass] work with
//     an interval given by its size.
//   - [Class.Invert], [Class.InvertAround], and [Class.InvertSwapping] invert
//     a class by an index, around an axis class, or by a pair of classes it
//     exchanges; [InversionFixedClasses] lists the classes an inversion
//     leaves in place.
//   - [Class.Multiply] multiplies a class; [Invertible],
//     [InverseMultiplier], [InvertibleMultipliers], and [Totient] describe
//     the invertible multipliers of N.
//   - [NewOperator], [Identity], [Transposition], [Inversion],
//     [InversionAround], [InversionSwapping], and [Multiplication] create
//     operators; [Operator.Apply], [Operator.Compose], and [Operator.Inverse]
//     apply, compose, and invert them.
//   - [Transpositions], [TranspositionsInversions], and [Operators] list the
//     groups of operators; [CountTranspositions],
//     [CountTranspositionsInversions], [CountOperators],
//     [CountTranspositionsInversionsBig], and [CountOperatorsBig] count them.
//   - [Class.String], [Parse], [Operator.String], and [ParseOperator] write
//     and read the canonical forms; [Class.Notation] and [ParseNotation]
//     write and read the literary notation of class numbers.
//
// Functions and methods whose names start with Must are wrappers for
// literals and values known to be valid: they return the same result as the
// operation without Must, or panic with its error. For data from users, use
// the operations that return errors.
//
// # Pitch classes
//
// A class is a pair of a number k and a number of positions N, with
// 0 <= k < N. It is the position of a note within the octave, without the
// octave: all C's of the traditional system are the class 0 mod 12. A class
// does not remember the system, the octave, the note name, or the pitch it
// comes from. Classes of different systems with the same N are equal, and
// classes of different N are never equal. The class of a note name or a
// step of a system is computed by the planned musex package system, not by
// this one.
//
// Classes are numbered by the integer notation of their system. In the
// traditional 12-position system C = 0, C♯ = D♭ = 1, D = 2, and so on to
// B = 11; this numbering belongs to the system, not to this package. The
// results of the operations depend on the choice of the class numbered 0:
// with a different zero, the same reflection has a different index.
//
// [Class.Compare] orders classes by N and then by number. This is an order
// of labels for sorting and stable output, not an order of pitch: the
// circle of classes has no beginning, and 11 mod 12 is not higher than
// 0 mod 12.
//
// # Intervals
//
// The ordered interval i<x, y> = (y - x) mod N is the distance from x upward
// to y around the circle, in the range 0 to N-1; the complementary interval
// is i<y, x>. The interval class ic(x, y) is the smaller of the two, in the
// range 0 to N/2. The interval class of a class with itself is 0; Open Music
// Theory (OMT) and MT21 name only the interval classes 1 to 6 for N = 12.
// The index sum (x + y) mod N is the index of the inversion that exchanges
// x and y.
//
// # Inversion
//
// The inversion In maps x to (n - x) mod N. In and TnI denote the same
// operation: invert around 0, then transpose by n; In is the shorter
// notation of Straus. The inversion I without an index is I0, that is,
// c.Invert(0), and I0 equals the multiplication by N-1. An inversion can
// also be given by an axis class, which stays in place
// ([Class.InvertAround]), or by a pair of classes that it exchanges
// ([Class.InvertSwapping]). For even N, In leaves two classes in place if n
// is even and none if n is odd; for odd N, exactly one
// ([InversionFixedClasses]).
//
// For N <= 2, every inversion is also a transposition, and for N = 1 there
// is a single class 0 mod 1 and a single operator.
//
// # Multiplication
//
// The multiplication Mm maps x to (m·x) mod N. If m is not invertible
// modulo N, that is, m and N have a common divisor greater than 1, Mm maps
// different classes to the same class; [Class.Multiply] accepts such m, but
// operators require an invertible m. For N = 12 the invertible multipliers
// are 1, 5, 7, and 11. M7 maps the chromatic circle of 12 classes onto the
// circle of fifths. For another N, the multiplier that gives a circle of
// fifths depends on the size of the fifth in the system and is not always
// invertible. M7 = I0∘M5, and M5 and M7 are each their own inverse.
//
// # Operators
//
// An operator is a map x ↦ m·x + n (mod N) with an invertible multiplier m.
// For N = 12 there are 48 operators: the twelve-tone operators (TTOs) of
// Morris, which combine transpositions, inversions, and the multiplications
// M5 and M7. In general there are N·φ(N) operators, where φ is Euler's
// totient ([Totient]). f.Compose(g) is the composition f∘g, which applies g
// first and then f. The canonical form "T<n>M<m> mod <N>" writes TnMm:
// multiply by m, then transpose by n.
//
// The literature defines these operations for N = 12, and for the diatonic
// 7 positions. Their generalization to any N in this package follows from
// the same formulas modulo N; it is a conclusion of the model of musex, not
// a statement of the literature.
//
// # Text
//
// [Class.String] returns the canonical form "<k> mod <N>", such as
// "1 mod 12", and [Operator.String] returns "T<n>M<m> mod <N>", such as
// "T5M7 mod 12". [Parse] and [ParseOperator] accept exactly these forms.
//
// [Class.Notation] writes the number of a class without N in a literary
// style. The styles differ only in the symbols for 10 and 11: [Decimal]
// writes 10 and 11, [TE] writes T and E (MT21, §33.3), [LowerTE] writes t and
// e (Whittall; OMT), and [AB] writes A and B (Morris). For N > 12 the
// literature has no one-character symbols, so only Decimal applies.
// [ParseNotation] reads a number strictly in the given style.
//
// # Large moduli
//
// N has no musical upper limit: any N from 1 to math.MaxInt is valid, and
// the limit is that of the int type. All arithmetic is exact for any int
// input, including math.MinInt and math.MaxInt. The lists [All],
// [Class.Cycle], [InvertibleMultipliers], [Transpositions],
// [TranspositionsInversions], and [Operators] are lazy iterators: their
// first elements are reached at once for any N, and a range over them can
// be stopped with break, but a full range over a very large N takes very
// long. Each range starts from the first element, and an iterator may be
// ranged over by several goroutines at once. A count that does not fit in an
// int returns [ErrOverflow]; the functions ending in Big return the exact
// value as a new [big.Int].
//
// # Errors
//
// Operations return errors that match exactly one of the following with
// [errors.Is]:
//
//   - [ErrInvalidClass]: a class cannot be created by [New] or [Reduce];
//   - [ErrInvalidModulus]: N is less than 1 in an operation that takes N;
//   - [ErrModulusMismatch]: two-place operations get classes or operators of
//     different N;
//   - [ErrNotInvertible]: a multiplier is not invertible;
//   - [ErrInvalidValue]: an operand is the zero Class or the zero Operator;
//   - [ErrNotation]: a text or a notation style is not valid;
//   - [ErrOverflow]: a count does not fit in an int.
//
// On error, operations return the zero Class or Operator, 0, false, "", or
// nil, and iterators that yield nothing (never a nil iterator). Only the
// functions and methods whose names start with Must panic, with the error
// as the panic value.
//
// An error of [Parse], [ParseNotation], or [ParseOperator] quotes the
// rejected text, unless it reports a modulus below 1, a style that is not
// valid or not defined for the modulus, or a multiplier that is not
// invertible. A text of up to 64 bytes is quoted in full. A longer text is
// quoted by at most its first 64 bytes, without splitting a valid UTF-8
// character, followed by its length in bytes. So the length of an error
// message of these functions is bounded however long the text is. The
// exact text of errors is not part of the API.
//
// # Invalid values
//
// The zero values Class{} and Operator{} are neither a class nor an
// operator: they mean that no value is given, and they are never treated as
// 0 mod 12 or the identity T0. IsValid reports false for them, Equal reports false
// even when both values are zero, and Compare orders them before every
// valid value (and reports 0 for a zero value with itself). Class{} ==
// Class{} is true, as for any comparable value; use Equal to compare
// classes. [Class.Number] returns -1 and [Class.Modulus] 0, [Operator.Modulus]
// returns 0 and [Operator.Multiplier] and [Operator.Index] -1. String returns
// "Class(invalid)" and "Operator(invalid)", which no parser accepts.
// [Class.Transpose], [Class.Invert], [Class.Multiply], and [Operator.Inverse]
// return the zero value, [Class.Cycle] yields nothing, and two-place
// operations, [InversionAround], and [Class.Notation] return
// [ErrInvalidValue].
//
// # Concurrency
//
// Classes, operators, and styles are immutable values, and the package has
// no mutable state. All functions, methods, and returned iterators are safe
// for concurrent use by multiple goroutines.
package pitchclass
