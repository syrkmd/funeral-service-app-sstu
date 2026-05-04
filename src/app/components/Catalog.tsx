import { useState } from "react";

const categories = ["Все", "Гробы", "Урны", "Цветы", "Памятные изделия"];

const products = [
  {
    id: 1,
    category: "Гробы",
    title: "Гроб дубовый премиум",
    description: "Массив дуба с бархатной внутренней отделкой",
    price: "320 000 ₽"
  },
  {
    id: 2,
    category: "Гробы",
    title: "Гроб из красного дерева",
    description: "Благородное красное дерево с атласной отделкой",
    price: "410 000 ₽"
  },
  {
    id: 3,
    category: "Гробы",
    title: "Гроб сосновый",
    description: "Простой и элегантный из натуральной сосны",
    price: "180 000 ₽"
  },
  {
    id: 4,
    category: "Урны",
    title: "Урна латунная",
    description: "Классическая латунная урна с возможностью гравировки",
    price: "45 000 ₽"
  },
  {
    id: 5,
    category: "Урны",
    title: "Урна керамическая",
    description: "Ручная работа, декоративная глазурь",
    price: "32 000 ₽"
  },
  {
    id: 6,
    category: "Урны",
    title: "Урна деревянная",
    description: "Орех ручной работы с бархатной отделкой",
    price: "38 000 ₽"
  },
  {
    id: 7,
    category: "Цветы",
    title: "Траурная композиция на гроб",
    description: "Полное покрытие гроба из роз и лилий",
    price: "45 000 ₽"
  },
  {
    id: 8,
    category: "Цветы",
    title: "Венок",
    description: "Элегантная круглая композиция на подставке",
    price: "28 000 ₽"
  },
  {
    id: 9,
    category: "Цветы",
    title: "Траурный букет",
    description: "Композиция из сезонных цветов в вазе",
    price: "12 000 ₽"
  },
  {
    id: 10,
    category: "Памятные изделия",
    title: "Книга памяти",
    description: "Книга для записей в кожаном переплёте",
    price: "7 500 ₽"
  },
  {
    id: 11,
    category: "Памятные изделия",
    title: "Памятные карточки",
    description: "Индивидуальная печать (комплект 100 шт)",
    price: "8 500 ₽"
  },
  {
    id: 12,
    category: "Памятные изделия",
    title: "Фотостенд",
    description: "Профессиональный фотоколлаж",
    price: "15 000 ₽"
  }
];

export function Catalog() {
  const [selectedCategory, setSelectedCategory] = useState("Все");

  const filteredProducts =
    selectedCategory === "Все"
      ? products
      : products.filter((p) => p.category === selectedCategory);

  return (
    <div className="max-w-7xl mx-auto px-6 py-12">
      <div className="mb-12 text-center">
        <h1 className="text-4xl mb-4 text-foreground">Каталог товаров</h1>
        <p className="text-lg text-muted-foreground">
          Качественные изделия для достойного прощания
        </p>
      </div>

      {/* Category Filter */}
      <div className="flex gap-3 mb-8 justify-center flex-wrap">
        {categories.map((category) => (
          <button
            key={category}
            onClick={() => setSelectedCategory(category)}
            className={`px-6 py-2 rounded-lg transition-colors ${
              selectedCategory === category
                ? "bg-primary text-primary-foreground"
                : "bg-secondary text-secondary-foreground hover:bg-accent"
            }`}
          >
            {category}
          </button>
        ))}
      </div>

      {/* Products Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredProducts.map((product) => (
          <div
            key={product.id}
            className="bg-card border border-border rounded-lg overflow-hidden hover:shadow-md transition-shadow"
          >
            <div className="bg-muted h-48 flex items-center justify-center">
              <span className="text-muted-foreground">{product.title}</span>
            </div>
            <div className="p-6">
              <div className="text-xs text-muted-foreground mb-2">
                {product.category}
              </div>
              <h3 className="mb-2 text-foreground">{product.title}</h3>
              <p className="text-sm text-muted-foreground mb-4">
                {product.description}
              </p>
              <div className="flex items-center justify-between">
                <span className="text-xl text-primary">{product.price}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
