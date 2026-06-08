export const placeholderImage = "/images/placeholder.png";

const productImages: Record<string, string> = {
  "Гроб дубовый": "/images/products/coffin-oak.png",
  "Гроб дубовый премиум": "/images/products/coffin-oak.png",
  "Гроб из красного дерева": "/images/products/coffin-mahogany.png",
  "Гроб сосновый": "/images/products/coffin-pine.png",
  "Урна": "/images/products/urn-brass.png",
  "Урна латунная": "/images/products/urn-brass.png",
  "Урна керамическая": "/images/products/urn-ceramic.png",
  "Урна деревянная": "/images/products/urn-ceramic.png",
  "Траурная композиция": "/images/products/casket-flowers.png",
  "Траурная композиция на гроб": "/images/products/casket-flowers.png",
  "Траурный букет": "/images/products/casket-flowers.png",
  "Венок": "/images/products/wreath.png",
  "Книга памяти": "/images/products/memory-book.png",
  "Памятные карточки": "/images/products/memory-book.png",
  "Фотостенд": "/images/products/memory-book.png",
};

const serviceImages: Record<string, string> = {
  "Традиционные похороны": "/images/services/traditional-funeral.png",
  "Кремация": "/images/services/cremation.png",
  "Поминальная церемония": "/images/services/memorial-ceremony.png",
  "Церемония на кладбище": "/images/services/cemetery-ceremony.png",
  "Простое погребение": "/images/services/cemetery-ceremony.png",
  "Прощание": "/images/services/farewell.png",
  "Транспортировка": "/images/services/transportation.png",
  "Бальзамирование": "/images/services/embalming.png",
  "Услуги некролога": "/images/services/obituary.png",
};

export function getProductImage(title: string) {
  return productImages[title.trim()] ?? placeholderImage;
}

export function getServiceImage(title: string) {
  return serviceImages[title.trim()] ?? placeholderImage;
}
