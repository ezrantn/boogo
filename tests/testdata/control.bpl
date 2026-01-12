procedure abs(x: int) returns (y: int)
{
  if (x < 0) {
    y := -x;
  } else {
    y := x;
  }
  return;
}

procedure main() {
  var r: int;
  call r := abs(-5);
  assert r == 5;
}