%%%%%// Compute the GIST feature for a single image


% Parameters:
param.imageSize = 128;
param.orientationsPerScale = [8 8 8 8];
param.numberBlocks = 4;
param.fc_prefilt = 4;
loc = 'D:/collections/formattedCollections/MMM-Export/crowdsourcedTestset/thumbnails/1147505.jpg';

    A = zeros(0, 512, 'double');

        rgb = imread(loc);
        rgb = single(rgb);
         imcell = 0.299 * rgb(:,:,1) + 0.587 * rgb(:,:,2) + 0.114 * rgb(:,:,3);

         gist = LMgist(imcell, '', param);
         gist
%    dlmwrite(strcat(loc2,'output',int2str(j),'.txt'), A);
