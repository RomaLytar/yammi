// Генерация lexorank позиции между двумя карточками
export function generatePosition(prev?: string, next?: string): string {
  if (!prev && !next) return 'm'
  if (!prev) return String.fromCharCode(next!.charCodeAt(0) - 1)
  if (!next) return prev + 'm'

  // Простая реализация: среднее между позициями
  const prevCode = prev.charCodeAt(prev.length - 1)
  const nextCode = next.charCodeAt(0)
  const mid = Math.floor((prevCode + nextCode) / 2)

  if (mid === prevCode) {
    return prev + 'm' // Если очень близко, добавляем букву
  }

  return String.fromCharCode(mid)
}
