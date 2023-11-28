import os
import subprocess
from dataclasses import dataclass
from typing import List

LOG_PATH = "./XX_BenchmarkOfbenchmarks"
@dataclass
class CallThing():
    drbenchmark: List[str]
    aut_args: List[str]
    aut_env: dict[str, str]
    name: str
@dataclass
class DrBenchmark():
    executable = "/root/aml-jens/bin/drbenchmark"
    perm_args = "-dev enx0c379652e326 -repeat 1 -callback /root/aml-jens/scripts/scream.sh".split(" ")
    def get(self) -> list:
        return [self.executable, *self.perm_args]
@dataclass
class BenchmarkRun():
    drbenchmark = DrBenchmark()
    benchmarks = [
        #{"addon_latency": 0, "scale": 1},
        #{"addon_latency": 0, "scale": 2},
        {"addon_latency": 20, "scale": 1},
        {"addon_latency": 20, "scale": 2},
        {"addon_latency": 40, "scale": 1},
        {"addon_latency": 40, "scale": 2},
        ]
    aut_changing_args = [
    [],
    ["-newcc"],
    ["-newcc", "-fincrease", "1"],
    ["-newcc", "-fincrease", "3"],
    ["-newcc", "-fincrease", "5"],
    ["-newcc", "-fincrease", "10"],    
    ]
    def getRuns(self) -> CallThing:
        env = os.environ.copy()
        for bm in self.benchmarks:
            latency = bm["addon_latency"]
            scale = bm["scale"]
            benchmark_file = create_benchmark(
                addonlatency= latency,
                scale= scale,
                base_path=os.path.join(LOG_PATH,"benchmarks"))
            for aut_args in self.aut_changing_args:
                env["BM_SCREAM_ARGS"] = " ".join(aut_args)
                env["BM_JSON_FILE"] = benchmark_file.split("/")[-1]
                scream_name = env["BM_SCREAM_ARGS"].replace(" ", "")
                if scream_name.find("newcc") == -1:
                    scream_name = "-oldcc"+scream_name
                scream_name = "scream"+scream_name    
                run_name = ".".join(env["BM_JSON_FILE"].split(".")[:-1])+ "_"+scream_name# remove spaces etc oldc c

                yield CallThing(
                    [*self.drbenchmark.get(), "-benchmark", benchmark_file],
                    aut_args,
                    aut_env = env,
                    name = run_name
                )
        

def main():
    run: CallThing
    for run in BenchmarkRun().getRuns():
        log_path = os.path.join(LOG_PATH, run.name)
        try:
            os.makedirs(log_path)
        except FileExistsError:
            pass
        env = run.aut_env.copy()
        env["LOG_PATH"] = log_path+"/"
        stdout = open(os.path.join(log_path, "drbenchmark.stdout"), "w")
        stderr = open(os.path.join(log_path, "drbenchmark.stderr"), "w")
        print(run.name, end=" ", flush=True)
        res = subprocess.run(
            [*run.drbenchmark, "-tag", run.name],
            stderr=stderr,
            stdout=stdout,
            env=env
        )
        print("🔴" if res.returncode != 0 else "🟢", flush=True)
        stdout.close()
        stderr.close()
                
def create_benchmark(addonlatency: int, scale: float, base_path="/etc/jens-cli/benchmarks/generated/") -> str:
    try:
        os.makedirs(base_path)
    except FileExistsError:
        pass
    path = os.path.join(base_path,f"WhiteBoard_{addonlatency}addon_{scale:.2f}scale.json")
    with open(path, "w") as fp:
        fp.write(f"""{{
  "Hash":"",
  "Inner":{{
    "Name":"WP{addonlatency}ms_x{scale:.2f}",
    "MaxBitrateEstimationTimeS":5,
    "Patterns":[
      {{
        "Path": "/etc/jens-cli/drp_3valleys.csv",
        "Hash": "020b6fc00d7a6f91c050a0833f11c18d"
      }},{{
        "Path": "/etc/jens-cli/drp_munich_autobahn.csv",
        "Hash": ""
      }},{{
        "Path": "/etc/jens-cli/drp_dominik_v1.csv",
        "Hash": ""
      }}
    ],
    "DrplaySetting":{{
      "DRP":{{
        "Frequency": 10,
        "Scale": {scale:.2f},
        "MinRateKbits": 1000,
        "WarmupBeforeDrpMs": 12000
      }},
      "TC": {{
        "Markfree": 4,
        "Markfull": 14,
        "Queuesizepackets": 10000,
        "Extralatency": {addonlatency}
      }}
    }}
  }}
}}
""")
    return path
if __name__ == '__main__':
    main()
