from example_package.example import add_one


def test_add_one():
    assert add_one(1) == 2


def test_add_one_negative():
    assert add_one(-1) == 0
