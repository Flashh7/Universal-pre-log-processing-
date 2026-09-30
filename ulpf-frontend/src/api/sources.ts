import type { Source } from "../types/source";
import { mockSources } from "../mock_data/sources";

export async function getSources(): Promise<Source[]> {
  return Promise.resolve(mockSources);
}