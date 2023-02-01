import unittest
from pathlib import Path

from tgbot_ping import get_runtime

project_root = Path(__file__).parent.parent
test_data = project_root / "tests" / "test_data"
inspects = list(test_data.joinpath("inspect").glob("*.json"))
stats = list(test_data.joinpath("stats").glob("*.json"))


class TestGetRuntime(unittest.TestCase):
    def test_runtime(self):
        display_name = "some bot"
        for inspect in inspects:
            for stat in stats:
                result = get_runtime("fake", display_name, test_data={"inspect": inspect, "stats": stat})
                self.assertIn(display_name, result)


if __name__ == '__main__':
    unittest.main()
