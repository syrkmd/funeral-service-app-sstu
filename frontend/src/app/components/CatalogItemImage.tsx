import {
  getProductImage,
  getServiceImage,
  placeholderImage,
} from "../utils/catalogImages";

type CatalogItemImageProps = {
  kind: "product" | "service";
  title: string;
  className?: string;
};

export function CatalogItemImage({
  kind,
  title,
  className = "",
}: CatalogItemImageProps) {
  const src = kind === "product"
    ? getProductImage(title)
    : getServiceImage(title);

  return (
    <img
      src={src}
      alt={title}
      loading="lazy"
      onError={(event) => {
        event.currentTarget.src = placeholderImage;
      }}
      className={className}
    />
  );
}
