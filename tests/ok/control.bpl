procedure abs(x: int) returns (y: int)
{
  if (x < 0) {
    y := -x;
  } else {
    y := x;
  }
  return;
}
