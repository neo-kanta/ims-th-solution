import { useState } from '#imports';

export const useGlobalProgress = () => {
  return useState<boolean>('globalProgress', () => false);
};
