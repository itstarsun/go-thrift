/* Comment */
// Comment
# Comment

enum TestEnum {
    ABC == /* ERROR "expected '=', found '=='" */ 1,
    XYZ = 2 (go.name == /* ERROR "expected '=', found '=='" */ "XYZ"),
}

namespace /* ERROR "expected definition, found 'namespace'" */ * test

struct TestStruct {
    required string 123 /* ERROR "expected 'IDENT', found 123" */,
    required list<string> strings == /* ERROR "expected '=', found '=='" */ []
}
