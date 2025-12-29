import dayjs from 'dayjs';

/**
 * Calcula a data da Páscoa usando o algoritmo de Meeus/Jones/Butcher
 */
const calculateEaster = (year) => {
  const a = year % 19;
  const b = Math.floor(year / 100);
  const c = year % 100;
  const d = Math.floor(b / 4);
  const e = b % 4;
  const f = Math.floor((b + 8) / 25);
  const g = Math.floor((b - f + 1) / 3);
  const h = (19 * a + b - d - g + 15) % 30;
  const i = Math.floor(c / 4);
  const k = c % 4;
  const l = (32 + 2 * e + 2 * i - h - k) % 7;
  const m = Math.floor((a + 11 * h + 22 * l) / 451);
  const month = Math.floor((h + l - 7 * m + 114) / 31);
  const day = ((h + l - 7 * m + 114) % 31) + 1;

  return dayjs(`${year}-${month.toString().padStart(2, '0')}-${day.toString().padStart(2, '0')}`);
};

/**
 * Retorna todos os feriados nacionais do Brasil para um determinado ano
 */
export const getBrazilianHolidays = (year) => {
  const easter = calculateEaster(year);

  const holidays = [
    // Feriados Fixos
    { date: dayjs(`${year}-01-01`), name: 'Confraternização Universal', type: 'fixed' },
    { date: dayjs(`${year}-04-21`), name: 'Tiradentes', type: 'fixed' },
    { date: dayjs(`${year}-05-01`), name: 'Dia do Trabalho', type: 'fixed' },
    { date: dayjs(`${year}-09-07`), name: 'Independência do Brasil', type: 'fixed' },
    { date: dayjs(`${year}-10-12`), name: 'Nossa Senhora Aparecida', type: 'fixed' },
    { date: dayjs(`${year}-11-02`), name: 'Finados', type: 'fixed' },
    { date: dayjs(`${year}-11-15`), name: 'Proclamação da República', type: 'fixed' },
    { date: dayjs(`${year}-12-25`), name: 'Natal', type: 'fixed' },

    // Feriados Móveis (baseados na Páscoa)
    { date: easter.subtract(47, 'day'), name: 'Carnaval', type: 'mobile' },
    { date: easter.subtract(2, 'day'), name: 'Sexta-feira Santa', type: 'mobile' },
    { date: easter, name: 'Páscoa', type: 'mobile' },
    { date: easter.add(60, 'day'), name: 'Corpus Christi', type: 'mobile' },
  ];

  return holidays;
};

/**
 * Verifica se uma data é feriado e retorna o nome
 */
export const getHolidayInfo = (date) => {
  const d = dayjs(date);
  const year = d.year();
  const holidays = getBrazilianHolidays(year);

  const holiday = holidays.find(h => h.date.format('YYYY-MM-DD') === d.format('YYYY-MM-DD'));

  return holiday || null;
};

/**
 * Verifica se uma data é feriado (boolean)
 */
export const isHoliday = (date) => {
  return getHolidayInfo(date) !== null;
};

/**
 * Retorna feriados para um range de datas (útil para semanas/meses)
 */
export const getHolidaysInRange = (startDate, endDate) => {
  const start = dayjs(startDate);
  const end = dayjs(endDate);
  const years = new Set();

  // Pega todos os anos no range
  let current = start;
  while (current.isBefore(end) || current.isSame(end, 'day')) {
    years.add(current.year());
    current = current.add(1, 'month');
  }

  // Coleta feriados de todos os anos relevantes
  const allHolidays = [];
  years.forEach(year => {
    allHolidays.push(...getBrazilianHolidays(year));
  });

  // Filtra apenas os feriados dentro do range
  return allHolidays.filter(h =>
    (h.date.isAfter(start) || h.date.isSame(start, 'day')) &&
    (h.date.isBefore(end) || h.date.isSame(end, 'day'))
  );
};
