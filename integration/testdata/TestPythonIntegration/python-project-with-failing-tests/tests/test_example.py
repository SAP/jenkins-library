from example_package.example import add_one


def test_add_one_fails():
    # deliberately wrong expectation
    assert add_one(1) == 99
