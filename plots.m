function main
    figure;
    hold on;
    % axis off;
    x1 = linspace(0,10,10);
    x1Extended = linspace(0,10,100);

    y1 = 5.*ones(length(x1),1);
    y2 = x1;
    y3 = (x1.^2)./5;
    y4 = (x1.^3)./25;
    y5 = exp(x1)./500;

    plot(x1,y1,'bo','DisplayName', 'VC');
    plot(x1,y1,'b-');
    
    plot(x1,y2,'rs','DisplayName', 'VL');
    plot(x1,y2,'r-');
    
    p = plot(x1,y3,'*','DisplayName', 'VQ');
    p.Color = "#EDB120";
    p = plot(x1Extended,interp1(x1, y3, x1Extended, 'spline'), '-');
    p.Color = "#EDB120";
    
    plot(x1,y4,'m^','DisplayName', 'VP');
    plot(x1Extended,interp1(x1, y4, x1Extended, 'spline'), 'm-');
    
    plot(x1,y5,'gd','DisplayName', 'VE');
    plot(x1Extended,interp1(x1, y5, x1Extended, 'spline'), 'g-');
    
    legend({'VC', '', 'VL', '', 'VQ', '', 'VP', '', 'VE', ''}, "FontSize", 11);
    axp = get(gca,'Position');
    xs=axp(1);
    xe=axp(1)+axp(3)+0.04;
    ys=axp(2);
    ye=axp(2)+axp(4)+0.05;
    annotation('arrow', [xs xe],[ys ys]);
    annotation('arrow', [xs xs],[ys ye]);
    set(gca,'XTick',[]);
    set(gca,'YTick',[]);
    text(10.2, -1.6, 'n', 'fontsize', 14);
    text(-1.4, 46, 'V_t\^(n)', 'fontsize', 14);
end
