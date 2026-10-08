/**
 * Dados de exemplo para desenvolver a interface sem o servidor Go.
 * Cada linha = uma criatura e o item de task que ela dropa.
 * (Substitua por chamadas reais em services/tasksApi.js.)
 */
export const MOCK_DROPS = [
  { criatura: 'Bashmu', item: 'Bashmu Fang' },
  { criatura: 'Bashmu', item: 'Bashmu Feather' },
  { criatura: 'Bashmu', item: 'Bashmu Tongue' },
  { criatura: 'Wyrm', item: 'Wyrm Scale' },
  { criatura: 'Dragon', item: 'Dragon Claw' },
  { criatura: 'Demon', item: 'Demon Horn' },
  { criatura: 'Hydra', item: 'Hydra Head' },
]

/** Exemplo exibido no botão "Usar exemplo" do analisador de log. */
export const MOCK_LOG = `Session data: From 2026-09-24, 22:37:41 to 2026-09-24, 22:52:05
Session: 00:14h
Loot: 251,888
Killed Monsters:
  37x bashmu
  52x juvenile bashmu
Looted Items:
  707x a platinum coin
  2x a violet gem
  1x a bashmu fang
  4x a bashmu tongue
  4x a bashmu feather`
