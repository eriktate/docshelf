export interface Options {
  retries: number; // number of times to attempt
  wait: number; // wait in milliseconds
}

const defaultOpts: Options = {
  retries: 10,
  wait: 20,
};

function sleep(ms: number): Promise<void> {
  return new Promise(r => setTimeout(r, ms));
}

async function retry(func: any, opts: Options = defaultOpts, _count: number = 0): Promise<void> {
  if (_count === opts.retries) {
    return;
  }

  for (let i = 0; i < opts.retries; i++) {
    try {
      await func();
      console.log("WORKED!");
      break;
    } catch (err) {
      console.log("FAILED!");
      await sleep(opts.wait);
    }
  }
}

export default retry;
