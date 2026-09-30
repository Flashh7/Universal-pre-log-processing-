import React from "react";

interface CardProps {
  title?: string;
  subtitle?: string;
  children: React.ReactNode;
  className?: string;
}

const Card = ({
  title,
  subtitle,
  children,
  className = "",
}: CardProps) => {
  return (
    <div className={`ulpf-card ${className}`}>
      {(title || subtitle) && (
        <div className="card-header">
          <div>
            {title && <h3 className="card-title">{title}</h3>}
            {subtitle && <p className="card-subtitle">{subtitle}</p>}
          </div>
        </div>
      )}

      <div className="card-content">
        {children}
      </div>
    </div>
  );
};

export default Card;