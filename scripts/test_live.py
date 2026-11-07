"""Offline production guard/sweeper regression tests; never use real credentials."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

SCRIPTS = Path(__file__).resolve().parent


class LiveGuardTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name)
        self.env = {
            "PATH": os.defpath,
            "ZEROPS_ENDPOINT": "https://api.app-prg1.zerops.io",
            "ZEROPS_TOKEN": "offline-fixture",
            "ZEROPS_ORGANIZATION_ID": "ci-org",
            "ZEROPS_CI_ORGANIZATION_ID": "ci-org",
            "ZEROPS_SANDBOX": "prod",
            "ZEROPS_PROD_LEDGER": str(self.base / "ledger"),
            "PULUMI_ZGUARD": str(self.base / "zguard"),
        }

    def run_script(self, name):
        return subprocess.run(["bash", str(SCRIPTS / name)], env=self.env,
                              text=True, capture_output=True, timeout=15)

    def test_valid_preflight(self):
        result = self.run_script("live-common.sh")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_missing_secret_names_are_reported_without_values(self):
        for key in ("ZEROPS_ENDPOINT", "ZEROPS_TOKEN", "ZEROPS_ORGANIZATION_ID", "ZEROPS_CI_ORGANIZATION_ID"):
            with self.subTest(key=key):
                original = self.env.pop(key)
                result = self.run_script("live-common.sh")
                self.env[key] = original
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(key, result.stderr)
                self.assertNotIn("offline-fixture", result.stdout + result.stderr)

    def test_configuration_fails_closed(self):
        for key, value in (
            ("ZEROPS_ENDPOINT", "https://api.app-prg1.zerops.io/"),
            ("ZEROPS_ENDPOINT", "https://foreign.invalid"),
            ("ZEROPS_CI_ORGANIZATION_ID", "customer-org"),
            ("ZEROPS_SANDBOX", "dev"),
            ("ZEROPS_PROD_LEDGER", "relative-ledger"),
            ("ZEROPS_TOKEN", " \t "),
        ):
            with self.subTest(key=key, value=value):
                original = self.env[key]
                self.env[key] = value
                result = self.run_script("live-common.sh")
                self.env[key] = original
                self.assertNotEqual(result.returncode, 0)

    def install_api_fixture(self, scenario):
        if not shutil.which("jq", path=self.env["PATH"]):
            self.skipTest("jq is required for sweeper tests")
        self.env.update(PATH=str(self.base) + os.pathsep + self.env["PATH"],
                        FIXTURE_SCENARIO=scenario, FIXTURE_DIR=str(self.base))
        curl = self.base / "curl"
        curl.write_text("""#!/usr/bin/env python3
import json, os, pathlib, sys
args = sys.argv[1:]
base = pathlib.Path(os.environ['FIXTURE_DIR'])
scenario = os.environ['FIXTURE_SCENARIO']
assert os.environ['ZEROPS_TOKEN'] not in ' '.join(args), 'token exposed in argv'
assert sys.stdin.read().strip() == 'Authorization: Bearer ' + os.environ['ZEROPS_TOKEN']
url = next(arg for arg in args if arg.startswith('https://'))
method = args[args.index('-X') + 1] if '-X' in args else 'GET'
target = {'id': 'owned', 'clientId': 'ci-org', 'name': 'tfacc-pulumi-fixture'}
ordinary = {'id': 'ordinary', 'clientId': 'ci-org', 'name': 'customer-app'}
if method == 'DELETE':
    raise AssertionError('production DELETE bypassed zguard')
elif url.endswith('/search'):
    if scenario == 'search-error':
        sys.exit(22)
    count_file = base / 'search-count'
    searches = int(count_file.read_text()) + 1 if count_file.exists() else 1
    count_file.write_text(str(searches))
    query = json.loads(args[args.index('-d') + 1])
    assert query['search'] == [{'name': 'clientId', 'operator': 'eq', 'value': 'ci-org'}]
    items = [ordinary]
    if not (base / 'deleted').exists():
        items.append(target)
    if scenario == 'late-leftover' and searches >= 3:
        items.append({'id': 'late', 'clientId': 'ci-org', 'name': 'tfacc-late'})
    if scenario == 'foreign-listing':
        items.append({'id': 'foreign', 'clientId': 'other-org', 'name': 'tfacc-foreign'})
    if scenario == 'pagination':
        items = [{'id': 'p' + str(i), 'clientId': 'ci-org', 'name': 'ordinary-' + str(i)} for i in range(100)] + items
    total = len(items)
    items = items[query['offset']:query['offset'] + query['limit']]
    print(json.dumps({'items': items, 'total': total}))
else:
    assert url.endswith('/owned')
    if scenario == 'ownership-changed':
        target['clientId'] = 'other-org'
    if scenario == 'prefix-changed':
        target['name'] = 'customer-app'
    if scenario == 'id-changed':
        target['id'] = 'different'
    print(json.dumps(target))
""")
        curl.chmod(0o700)
        # zguard is an external production dependency, mocked only in this offline
        # test harness. Production has no fallback when the binary is unavailable.
        guard = self.base / "zguard"
        guard.write_text("""#!/usr/bin/env python3
import json, os, pathlib, sys
base = pathlib.Path(os.environ['FIXTURE_DIR'])
scenario = os.environ['FIXTURE_SCENARIO']
args = sys.argv[1:]
assert os.environ['ZEROPS_TOKEN'] not in ' '.join(args)
if args[0] == 'record-project':
    assert args == ['record-project', 'owned', 'tfacc-pulumi-fixture', 'probe']
    if scenario == 'adoption-rejected':
        sys.exit(97)
    with (base / 'adoptions').open('a') as stream:
        stream.write(json.dumps(args[1:]) + '\\n')
    pathlib.Path(os.environ['ZEROPS_PROD_LEDGER']).write_text('owned\\n')
elif args[0] == 'request':
    assert args == ['request', 'DELETE', '/project/owned']
    assert pathlib.Path(os.environ['ZEROPS_PROD_LEDGER']).read_text() == 'owned\\n'
    if scenario == 'guard-rejected':
        sys.exit(97)
    if scenario == 'delete-error':
        print(json.dumps({'status': 403, 'body': '{}', 'headers': '', 'time': 0}))
        sys.exit(0)
    with (base / 'deletes').open('a') as stream:
        stream.write(os.environ['ZEROPS_ENDPOINT'] + '/api/rest/public/project/owned\\n')
    (base / 'deleted').touch()
    print(json.dumps({'status': 200, 'body': '{}', 'headers': '', 'time': 0}))
else:
    raise AssertionError('unexpected guard command')
""")
        guard.chmod(0o700)

    def test_sweep_deletes_only_owned_prefix_and_confirms_zero(self):
        for scenario in ("normal", "pagination"):
            with self.subTest(scenario=scenario):
                for name in ("deleted", "deletes", "adoptions", "ledger"):
                    (self.base / name).unlink(missing_ok=True)
                self.install_api_fixture(scenario)
                result = self.run_script("live-sweep.sh")
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn("zero tfacc- projects", result.stdout)
                self.assertEqual((self.base / "deletes").read_text().splitlines(),
                                 [self.env["ZEROPS_ENDPOINT"] + "/api/rest/public/project/owned"])
                self.assertEqual([json.loads(line) for line in (self.base / "adoptions").read_text().splitlines()],
                                 [["owned", "tfacc-pulumi-fixture", "probe"]])
                self.assertNotIn("offline-fixture", result.stdout + result.stderr)

    def test_sweep_refuses_foreign_or_changed_ownership_and_api_errors(self):
        for scenario in ("foreign-listing", "ownership-changed", "prefix-changed", "id-changed", "search-error"):
            with self.subTest(scenario=scenario):
                self.install_api_fixture(scenario)
                result = self.run_script("live-sweep.sh")
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse((self.base / "deletes").exists())
                self.assertFalse((self.base / "adoptions").exists())
                self.assertFalse((self.base / "ledger").exists())

    def test_guard_rejection_or_delete_failure_stops_cleanup(self):
        for scenario in ("adoption-rejected", "guard-rejected", "delete-error"):
            with self.subTest(scenario=scenario):
                self.install_api_fixture(scenario)
                result = self.run_script("live-sweep.sh")
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse((self.base / "deletes").exists())
                self.assertFalse((self.base / "deleted").exists())
                self.assertNotIn("offline-fixture", result.stdout + result.stderr)

    def test_sweep_requires_absolute_executable_guard(self):
        self.install_api_fixture("normal")
        for path in ("", "relative-zguard", str(self.base / "missing-guard")):
            with self.subTest(path=path):
                self.env["PULUMI_ZGUARD"] = path
                result = self.run_script("live-sweep.sh")
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse((self.base / "search-count").exists())
                self.assertFalse((self.base / "adoptions").exists())
                self.assertFalse((self.base / "deletes").exists())

    def test_sweep_preflight_rejects_wrong_org_before_inventory_or_adoption(self):
        self.install_api_fixture("normal")
        self.env["ZEROPS_CI_ORGANIZATION_ID"] = "other-org"
        result = self.run_script("live-sweep.sh")
        self.assertNotEqual(result.returncode, 0)
        for name in ("search-count", "adoptions", "ledger", "deletes"):
            self.assertFalse((self.base / name).exists())

    def test_sweep_fails_when_final_listing_has_leftovers(self):
        self.install_api_fixture("late-leftover")
        result = self.run_script("live-sweep.sh")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("tfacc- projects remain", result.stderr)


if __name__ == "__main__":
    unittest.main()
