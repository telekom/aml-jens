
data = ["A", "B", "C"]
end = 0
i = 0
operator = 1
while True:
    print(f"[{end}]: {abs(i)} - {data[min(abs(i), 2)]} ")
    i += operator
    if i >= 3 or i < 0:
        operator *= -1
        i += operator
    end += 1
    if end > 10:
        break
