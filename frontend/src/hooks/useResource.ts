import { useEffect, useState } from "react";
export function useResource<T>(load: (signal: AbortSignal) => Promise<T>) {
  const [attempt, setAttempt] = useState(0);
  const [state, setState] = useState<{
    data?: T;
    error?: Error;
    load: typeof load;
    attempt: number;
  }>();
  useEffect(() => {
    const controller = new AbortController();
    load(controller.signal)
      .then((data) => {
        if (!controller.signal.aborted) setState({ data, load, attempt });
      })
      .catch((error) => {
        if (!controller.signal.aborted) setState({ error, load, attempt });
      });
    return () => controller.abort();
  }, [load, attempt]);
  const current =
    state?.load === load && state?.attempt === attempt ? state : undefined;
  return {
    data: current?.data,
    error: current?.error,
    loading: !current,
    retry: () => setAttempt((value) => value + 1),
  };
}
