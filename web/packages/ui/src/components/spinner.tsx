import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { cva, type VariantProps } from 'class-variance-authority';
import {
  Loader2Icon,
  LoaderCircleIcon,
  LoaderIcon,
  LoaderPinwheelIcon,
  type LucideProps
} from 'lucide-react';

const spinnerVariant = cva('animate-spin', {
  variants: {
    size: {
      xs: 'size-2',
      sm: 'size-4',
      md: 'size-6',
      lg: 'size-8',
      xl: 'size-10',
      '2xl': 'size-12'
    }
  },
  defaultVariants: {
    size: 'md'
  }
});

const sizeToPixels: Record<NonNullable<VariantProps<typeof spinnerVariant>['size']>, number> = {
  xs: 8,
  sm: 16,
  md: 24,
  lg: 32,
  xl: 40,
  '2xl': 48
};

export type SpinnerProps = React.ComponentProps<'svg'> &
  VariantProps<typeof spinnerVariant> &
  LucideProps & {
    variant?:
      | 'default'
      | 'throbber'
      | 'pinwheel'
      | 'circle-filled'
      | 'ellipsis'
      | 'ring'
      | 'bars'
      | 'infinite';
  };

type SpinnerVariantProps = Omit<SpinnerProps, 'variant'>;

const Throbber = ({ className, size, ...props }: SpinnerVariantProps) => (
  <LoaderIcon
    className={cn(spinnerVariant({ size }), className)}
    {...props}
  />
);

const Pinwheel = ({ className, size, ...props }: SpinnerVariantProps) => (
  <LoaderPinwheelIcon
    className={cn(spinnerVariant({ size }), className)}
    {...props}
  />
);

const CircleFilled = ({ className, size, ...props }: SpinnerVariantProps) => {
  const px = sizeToPixels[size ?? 'md'];
  return (
    <div
      className="relative"
      style={{ width: px, height: px }}
    >
      <div className="absolute inset-0 rotate-180">
        <LoaderCircleIcon
          className={cn('animate-spin opacity-20', className)}
          size={px}
          {...props}
        />
      </div>
      <LoaderCircleIcon
        className={cn('relative animate-spin', className)}
        size={px}
        {...props}
      />
    </div>
  );
};

const Ellipsis = ({ size, className, ...props }: SpinnerVariantProps) => {
  const px = sizeToPixels[size ?? 'md'];
  return (
    <svg
      height={px}
      viewBox="0 24"
      width={px}
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <title>Loading...</title>
      <circle
        cx="4"
        cy="12"
        fill="currentColor"
        r="2"
      >
        <animate
          attributeName="cy"
          begin="0;ellipsis3.end+0.25s"
          calcMode="spline"
          dur="0.6s"
          id="ellipsis1"
          keySplines=".33,.66,.66,1;.33,0,.66,.33"
          values="12;6;12"
        />
      </circle>
      <circle
        cx="12"
        cy="12"
        fill="currentColor"
        r="2"
      >
        <animate
          attributeName="cy"
          begin="ellipsis1.begin+0.1s"
          calcMode="spline"
          dur="0.6s"
          keySplines=".33,.66,.66,1;.33,0,.66,.33"
          values="12;6;12"
        />
      </circle>
      <circle
        cx="20"
        cy="12"
        fill="currentColor"
        r="2"
      >
        <animate
          attributeName="cy"
          begin="ellipsis1.begin+0.2s"
          calcMode="spline"
          dur="0.6s"
          id="ellipsis3"
          keySplines=".33,.66,.66,1;.33,0,.66,.33"
          values="12;6;12"
        />
      </circle>
    </svg>
  );
};

const Ring = ({ size, className, ...props }: SpinnerVariantProps) => {
  const px = sizeToPixels[size ?? 'md'];
  return (
    <svg
      height={px}
      stroke="currentColor"
      viewBox="0 44"
      width={px}
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <title>Loading...</title>
      <g
        fill="none"
        fillRule="evenodd"
        strokeWidth="2"
      >
        <circle
          cx="22"
          cy="22"
          r="1"
        >
          <animate
            attributeName="r"
            begin="0s"
            calcMode="spline"
            dur="1.8s"
            keySplines="0.165, 0.84, 0.44, 1"
            keyTimes="0; 1"
            repeatCount="indefinite"
            values="1; 20"
          />
          <animate
            attributeName="stroke-opacity"
            begin="0s"
            calcMode="spline"
            dur="1.8s"
            keySplines="0.3, 0.61, 0.355, 1"
            keyTimes="0; 1"
            repeatCount="indefinite"
            values="1; 0"
          />
        </circle>
        <circle
          cx="22"
          cy="22"
          r="1"
        >
          <animate
            attributeName="r"
            begin="-0.9s"
            calcMode="spline"
            dur="1.8s"
            keySplines="0.165, 0.84, 0.44, 1"
            keyTimes="0; 1"
            repeatCount="indefinite"
            values="1; 20"
          />
          <animate
            attributeName="stroke-opacity"
            begin="-0.9s"
            calcMode="spline"
            dur="1.8s"
            keySplines="0.3, 0.61, 0.355, 1"
            keyTimes="0; 1"
            repeatCount="indefinite"
            values="1; 0"
          />
        </circle>
      </g>
    </svg>
  );
};

const Bars = ({ size, className, ...props }: SpinnerVariantProps) => {
  const px = sizeToPixels[size ?? 'md'];
  return (
    <svg
      height={px}
      viewBox="0 24"
      width={px}
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <title>Loading...</title>
      <style>{`
      .spinner-bar {
        animation: spinner-bars-animation .8s linear infinite;
        animation-delay: -.8s;
      }
      .spinner-bars-2 {
        animation-delay: -.65s;
      }
      .spinner-bars-3 {
        animation-delay: -0.5s;
      }
      @keyframes spinner-bars-animation {
        0% {
          y: 1px;
          height: 22px;
        }
        93.75% {
          y: 5px;
          height: 14px;
          opacity: 0.2;
        }
      }
    `}</style>
      <rect
        className="spinner-bar"
        fill="currentColor"
        height="22"
        width="6"
        x="1"
        y="1"
      />
      <rect
        className="spinner-bar spinner-bars-2"
        fill="currentColor"
        height="22"
        width="6"
        x="9"
        y="1"
      />
      <rect
        className="spinner-bar spinner-bars-3"
        fill="currentColor"
        height="22"
        width="6"
        x="17"
        y="1"
      />
    </svg>
  );
};

const Infinite = ({ size, className, ...props }: SpinnerVariantProps) => {
  const px = sizeToPixels[size ?? 'md'];
  return (
    <svg
      height={px}
      preserveAspectRatio="xMidYMid"
      viewBox="0 100"
      width={px}
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <title>Loading...</title>
      <path
        d="M24.3 30C11.4 30 5 43.3 50s6.4 20 19.3 20c19.3 0 32.1-40 51.4-40 C88.6 95 50s-6.4 20-19.3 20C56.4 70 43.6 24.3 30z"
        fill="none"
        stroke="currentColor"
        strokeDasharray="205.271142578125 51.317785644531256"
        strokeLinecap="round"
        strokeWidth="10"
        style={{
          transform: 'scale(0.8)',
          transformOrigin: '50px'
        }}
      >
        <animate
          attributeName="stroke-dashoffset"
          dur="2s"
          keyTimes="0;1"
          repeatCount="indefinite"
          values="0;256.58892822265625"
        />
      </path>
    </svg>
  );
};

export const Spinner = ({ variant, size, className, ...props }: SpinnerProps) => {
  switch (variant) {
    case 'throbber':
      return (
        <Throbber
          size={size}
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
    case 'pinwheel':
      return (
        <Pinwheel
          size={size}
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
    case 'circle-filled':
      return (
        <CircleFilled
          size={size}
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
    case 'ellipsis':
      return (
        <Ellipsis
          size={size}
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
    case 'ring':
      return (
        <Ring
          size={size}
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
    case 'bars':
      return (
        <Bars
          size={size}
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
    case 'infinite':
      return (
        <Infinite
          size={size}
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
    default:
      return (
        <Loader2Icon
          className={cn(spinnerVariant({ size }), className)}
          {...props}
        />
      );
  }
};
